package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"gopkg.in/yaml.v3"
)

type OutputType string

const (
	JSON       OutputType = "json"
	PrettyJSON OutputType = "pretty-json"
	YAML       OutputType = "yaml"
	Table      OutputType = "table"
)

func FormatAs(s string) OutputType {
	switch strings.ToLower(s) {
	case "json":
		return JSON
	case "pretty-json":
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
	case map[any]any:
		for k, v := range tData {
			fmt.Fprintf(out, "%v: -----------\n", k)
			err = listAsTable(out, v)
			if err != nil {
				return err
			}
		}
		fmt.Fprintln(out, "")
	case []any:
		for _, v := range tData {
			err = listAsTable(out, v)
			if err != nil {
				return err
			}
		}
	case *any:
		err = PrintTable(out, *tData)
	case any:
		err = printTableAny(out, tData)
	default:
		err = fmt.Errorf("unknown type: %T", tData)
	}

	return err
}

func printTableAny(out io.Writer, data any) (err error) {
	handledOnSwitch := true
	switch tData := data.(type) {
	case tsuru.App:
		_, err = fmt.Fprintf(out, "Handled!!!! %v\n\n", tData.Name) // XXX: fix this
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
		switch tData := reflect.ValueOf(data).Field(i).Interface().(type) {
		case nil:
		case []byte:
			simpleInfo[field.Name] = string(tData)
		case string:
			if tData != "" {
				simpleInfo[field.Name] = tData
			}
		case int, int16, int32, int64, int8, uint, uint16, uint32, uint64, uint8, float32, float64, complex64, complex128, bool:
			simpleInfo[field.Name] = tData
		case []string:
			simpleInfo[field.Name] = strings.Join(tData, ", ")
		case map[any]any:
			complexInfo[field.Name] = tData // XXX: fix this
		case []any:
			complexInfo[field.Name] = []any{}
			for _, v := range tData {
				complexInfo[field.Name] = append(complexInfo[field.Name].([]any), v)
			}
		default:
			complexInfo[field.Name] = tData
		}

		// inter := reflect.ValueOf(data).Field(i).Interface()
		// _, err = fmt.Fprintf(out, "%s (%T):\t%v\n", field.Name, inter, inter)
	}

	// print simple info
	fmt.Fprintf(out, "============= simpleInfo: ==============\n")
	for _, k := range sortedKeys(simpleInfo) {
		_, err = fmt.Fprintf(out, "%s:\t%v\n", k, simpleInfo[k])
		if err != nil {
			return err
		}
	}

	// print complex info
	// XXX: sort keys
	fmt.Fprintf(out, "\n============= complexInfo: ==============\n")
	for _, k := range sortedKeys(complexInfo) {
		_, err = fmt.Fprintf(out, "\n%s:\n\t%v\n", k, complexInfo[k])
		if err != nil {
			return err
		}
	}
	return err
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
