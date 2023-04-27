// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

type CustomFieldFunc func(any) string

type TableViewOptions struct {
	// ShowFields is a list of fields to exclusively show in the table.
	ShowFields []string
	// HideFields is a list of fields to hide in the table.
	// If ShowFields is not empty, this list will be ignored.
	HiddenFields []string
	// CustomFieldFunc is a map of field name to a function that returns a string to show in the table.
	CustomFieldFunc map[string]CustomFieldFunc
}

func (o *TableViewOptions) isFieldVisible(field string) bool {
	if o == nil {
		return true
	}
	if len(o.ShowFields) > 0 {
		return Contains(o.ShowFields, field)
	}
	if len(o.HiddenFields) > 0 {
		return !Contains(o.HiddenFields, field)
	}
	return true
}

func (o *TableViewOptions) visibleFieldsFromMap(m map[string]any) []string {
	if o == nil {
		return sortedKeys(m)
	}
	if len(o.ShowFields) > 0 {
		return o.ShowFields
	}
	return sortedKeysExcept(m, o.HiddenFields)
}
func (o *TableViewOptions) visibleFieldsFromSlice(ss []string) []string {
	ret := make([]string, 0, len(ss))
	for _, s := range ss {
		if o.isFieldVisible(s) {
			ret = append(ret, s)
		}
	}
	sort.Strings(ret)
	return ret
}

// PrintTable prints the data to out in a table format.
// If data is a slice, it will print each element in a sub-table.
func PrintTable(out io.Writer, data any, opts *TableViewOptions) (err error) {
	w := tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
	defer w.Flush()
	return printTable(w, data, opts)
}

// PrintTableList prints the data to out in a single table format (slice fields may be ignored).
func PrintTableList(out io.Writer, data any, opts *TableViewOptions) error {
	w := tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
	defer w.Flush()
	return printTableList(w, data, opts)
}

func printTable(out io.Writer, data any, opts *TableViewOptions) (err error) {
	if data == nil {
		return nil
	}

	switch tData := data.(type) {
	case []byte:
		_, err = out.Write(tData)
	case string:
		_, err = fmt.Fprintln(out, tData)
	case int, int16, int32, int64, int8, uint, uint16, uint32, uint64, uint8, float32, float64, complex64, complex128, bool:
		_, err = fmt.Fprintln(out, tData)
	case []string:
		_, err = fmt.Println(strings.Join(tData, "\n"))
	case io.Reader:
		_, err = io.Copy(out, tData)
	case map[string]any:
		for _, k := range opts.visibleFieldsFromMap(tData) {
			fmt.Fprintf(out, "%v: -----------\n", k)
			err = printTableList(out, tData[k], opts)
			if err != nil {
				return err // TODO: return a multi-error
			}
		}
		fmt.Fprintln(out, "")
	case []any:
		for _, v := range tData {
			err = printTableList(out, v, opts)
			if err != nil {
				return err
			}
		}
	case *any:
		err = PrintTable(out, *tData, opts)
	case any:
		err = printTableAny(out, tData, opts)
	default:
		err = fmt.Errorf("unknown type: %T", tData)
	}

	return err
}

func printTableAny(out io.Writer, data any, opts *TableViewOptions) (err error) {
	handledOnSwitch := true
	switch tData := data.(type) {
	case tsuru.App:
		_, err = fmt.Fprintf(io.Discard, "Handled!!!! %v\n\n", tData.Name) // XXX: fix this
		handledOnSwitch = false
	default:
		handledOnSwitch = false
	}
	if handledOnSwitch {
		return err
	}

	simpleInfo := map[string]any{}
	complexInfo := map[string]any{}

	// No custom printer found, try to print as best as we can
	dt := reflect.TypeOf(data)
	for _, field := range reflect.VisibleFields(dt) {
		if !opts.isFieldVisible(field.Name) {
			continue
		}

		v := reflect.ValueOf(data).FieldByName(field.Name)
		kind := v.Kind()
		switch kind {
		case reflect.Bool:
			simpleInfo[field.Name] = v.Bool()
		case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
			simpleInfo[field.Name] = v.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
			simpleInfo[field.Name] = v.Uint()
		case reflect.Float32, reflect.Float64:
			simpleInfo[field.Name] = v.Float()
		case reflect.String:
			simpleInfo[field.Name] = v.String()
		case reflect.Slice:
			if tData, ok := v.Interface().([]string); ok {
				simpleInfo[field.Name] = strings.Join(tData, ", ")
			} else {
				if v.Len() == 0 {
					continue
				}
				complexInfo[field.Name] = v.Interface()
			}
		case reflect.Map:
			if v.Len() == 0 {
				continue
			}
			fmt.Printf("map: %v\n", v.Interface()) // XXX: implement this
		case reflect.Chan:
		default:
			complexInfo[field.Name] = v.Interface()
		}
	}

	// print simple info
	for _, k := range sortedKeys(simpleInfo) {
		_, err = fmt.Fprintf(out, "%s:\t%v\n", k, simpleInfo[k])
		if err != nil {
			return err
		}
	}

	// print complex info
	for _, k := range sortedKeys(complexInfo) {
		fmt.Fprintf(out, "\n%s (%T):\n", k, complexInfo[k])
		if err = printTableList(out, complexInfo[k], opts); err != nil {
			return err
		}
	}
	return err
}

func printTableList(out io.Writer, data any, opts *TableViewOptions) (err error) {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		switch reflect.TypeOf(data).Elem().Kind() {
		case reflect.String:
			_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(data.([]string), "\n\t"))
			return err
		case reflect.Map:
			return fmt.Errorf("not implemented: printSubTable(%T)", data)
		case reflect.Struct:
			err = printTableListOfStructs(out, data, opts)
			fmt.Fprintln(out, "")
			return err
		default:
			return fmt.Errorf("slice with unknown type: %T", data)
		}
	case reflect.Map:
		return fmt.Errorf("not implemented: printSubTable(%T)", data)
	case reflect.Struct:
		err = printTableOfStructs(out, data, opts)
		fmt.Fprintln(out, "")
		return err
	default:
		return fmt.Errorf("unknown type: %T", data)
	}
}

func UpperCase(ss []string) []string {
	ret := make([]string, len(ss))
	for i, s := range ss {
		ret[i] = strings.ToUpper(s)
	}
	return ret
}

func printTableOfStructs(out io.Writer, data any, opts *TableViewOptions) (err error) {
	keys := []string{}
	for _, vf := range reflect.VisibleFields(reflect.TypeOf(data)) {
		keys = append(keys, vf.Name)
	}
	keys = opts.visibleFieldsFromSlice(keys)

	sort.Strings(keys) // XXX: sort with defaults first?
	_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(UpperCase(keys), "\t"))
	for _, k := range keys {
		item := reflect.ValueOf(data).FieldByName(k).Interface()
		if fn, ok := opts.CustomFieldFunc[k]; ok {
			fmt.Fprintf(out, "\t%s", fn(item))
		} else {
			fmt.Fprintf(out, "\t%v", item)
		}
	}
	return
}

func printTableListOfStructs(out io.Writer, data any, opts *TableViewOptions) (err error) {
	keys := []string{}
	for _, vf := range reflect.VisibleFields(reflect.TypeOf(data).Elem()) {
		keys = append(keys, vf.Name)
	}
	keys = opts.visibleFieldsFromSlice(keys)

	_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(UpperCase(keys), "\t"))
	for i := 0; i < reflect.ValueOf(data).Len(); i++ {
		item := reflect.ValueOf(data).Index(i).Interface()
		for _, k := range keys {
			if fn, ok := opts.CustomFieldFunc[k]; ok {
				fmt.Fprintf(out, "\t%s", fn(item))
			} else {
			fmt.Fprintf(out, "\t%v", reflect.ValueOf(item).FieldByName(k).Interface())
			}
		}
		fmt.Fprintln(out, "")
	}
	return
}

func sortedKeys(d map[string]any) []string {
	// XXX: get some fields first (Name, Description, ID, etc...)
	sortedKeys := make([]string, 0, len(d))
	for k := range d {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func sortedKeysExcept(d map[string]any, exceptions []string) []string {
	// XXX: get some fields first (Name, Description, ID, etc...)
	sortedKeys := make([]string, 0, len(d))
	for k := range d {
		if Contains(exceptions, k) {
			continue
		}
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
