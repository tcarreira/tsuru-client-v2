package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
		out.Write(tData)
	case string:
		fmt.Fprintln(out, tData)
	case io.Reader:
		_, err = io.Copy(out, tData)
	case map[string]any:
		for k, v := range tData {
			fmt.Fprintf(out, "%s: %v", k, v)
		}
	}

	return err
}
