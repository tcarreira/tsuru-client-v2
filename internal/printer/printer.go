package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"gopkg.in/yaml.v3"
)

type OutputType string

const (
	// every OutputType should be mapped inside PrintInfo()
	JSON       OutputType = "JSON"
	PrettyJSON OutputType = "PrettyJSON"
	YAML       OutputType = "YAML"
	Table      OutputType = "Table"
)

func FormatAs(s string) OutputType {
	switch strings.ToLower(s) {
	case "json":
		return JSON
	case "pretty-json", "prettyjson":
		return PrettyJSON
	case "yaml":
		return YAML
	case "table":
		return Table
	default:
		return Table
	}
}

func PrintJSON(out io.Writer, data any) error {
	if data == nil {
		return nil
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error converting to json: %w", err)
	}
	fmt.Fprintln(out, string(dataByte))
	return nil
}

func PrintPrettyJSON(out io.Writer, data any) error {
	if data == nil {
		return nil
	}
	dataByte, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error converting to json: %w", err)
	}
	fmt.Fprintln(out, string(dataByte))
	return nil
}

func PrintYAML(out io.Writer, data any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// yaml.v3 panics a lot: https://github.com/go-yaml/yaml/issues/954
			err = fmt.Errorf("error converting to yaml (panic): %v", r)
		}
	}()

	if data == nil {
		return nil
	}
	dataByte, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error converting to yaml: %w", err)
	}
	_, err = out.Write(dataByte)
	return err
}

func PrintTable(out io.Writer, data any) (err error) {
	w := tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
	defer w.Flush()

	if data == nil {
		return nil
	}

	switch tData := data.(type) {
	case []byte:
		_, err = w.Write(tData)
	case string:
		_, err = fmt.Fprintln(w, tData)
	case int, int16, int32, int64, int8, uint, uint16, uint32, uint64, uint8, float32, float64, complex64, complex128, bool:
		_, err = fmt.Fprintln(w, tData)
	case []string:
		_, err = fmt.Println(strings.Join(tData, "\n"))
	case io.Reader:
		_, err = io.Copy(w, tData)
	case map[any]any:
		for k, v := range tData {
			fmt.Fprintf(w, "%v: -----------\n", k)
			err = listAsTable(w, v)
			if err != nil {
				return err
			}
		}
		fmt.Fprintln(w, "")
	case []any:
		for _, v := range tData {
			err = listAsTable(w, v)
			if err != nil {
				return err
			}
		}
	case *any:
		err = PrintTable(w, *tData)
	case any:
		err = printTableAny(w, tData)
	default:
		err = fmt.Errorf("unknown type: %T", tData)
	}

	return err
}

func printTableAny(out io.Writer, data any) (err error) {
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
	for i := 0; i < dt.NumField(); i++ {
		field := dt.Field(i)

		v := reflect.ValueOf(data).Field(i)
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
			fmt.Printf("map: %v\n", v.Interface())
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
		if err = printSubTable(out, complexInfo[k]); err != nil {
			return err
		}
	}
	return err
}

func printSubTable(out io.Writer, data any) (err error) {
	keys := []string{}
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		switch reflect.TypeOf(data).Elem().Kind() {
		case reflect.String:
			_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(data.([]string), "\n\t"))
			return err
		case reflect.Map:
			for _, v := range data.([]map[string]any) {
				for k := range v {
					keys = append(keys, k)
				}
			}
		case reflect.Struct:
			return printSubTableOfStructs(out, data)
		default:
			return fmt.Errorf("unknown type: %T", data)

		}
	}

	sort.Strings(keys)
	_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(keys, "\t"))
	for _, k := range keys {
		_, err = fmt.Fprintf(out, "\t%s:\t%v\n", k, data.(map[any]any)[k])
	}

	return
}

func printSubTableOfStructs(out io.Writer, data any) (err error) {
	keys := []string{}
	for _, vf := range reflect.VisibleFields(reflect.TypeOf(data).Elem()) {
		keys = append(keys, vf.Name)
	}

	sort.Strings(keys) // XXX: sort with defaults first?
	_, err = fmt.Fprintf(out, "\t%s\n", strings.Join(keys, "\t"))
	for i := 0; i < reflect.ValueOf(data).Len(); i++ {
		item := reflect.ValueOf(data).Index(i).Interface()
		for _, k := range keys {
			fmt.Fprintf(out, "\t%v", reflect.ValueOf(item).FieldByName(k).Interface())
		}
		fmt.Fprintln(out, "")
	}
	return
}

func sortedKeys(d map[string]any) []string {
	// XXX: get some fields first (Name, Description, ID, etc...)
	sortedKeys := []string{}
	for k := range d {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func listAsTable(out io.Writer, data any) (err error) {
	return fmt.Errorf("not implemented")
}
