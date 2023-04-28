// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

type FieldType struct {
	Name  string
	Value any
}

type ListType struct {
	Name    string
	Headers []string
	Values  [][]any
}

type PrintableType struct {
	SimpleFields []FieldType
	ComplexField []ListType
}

var _ json.Marshaler = &PrintableType{}
var _ json.Marshaler = &ListType{}
var _ yaml.Marshaler = &PrintableType{}
var _ yaml.Marshaler = &ListType{}

func (p *PrintableType) MarshalJSON() ([]byte, error) {
	ret := make(map[string]any)
	for _, f := range p.SimpleFields {
		ret[f.Name] = f.Value
	}

	for _, f := range p.ComplexField {
		ret[f.Name] = f.ToSliceOfMap()
	}
	return json.Marshal(p.SimpleFields)
}

func (l *ListType) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.ToSliceOfMap())
}

func (l *ListType) ToSliceOfMap() []map[string]any {
	ret := []map[string]any{}
	for i, v := range l.Values {
		ret = append(ret, map[string]any{})
		for j, h := range l.Headers {
			ret[i][h] = v[j]
		}
	}
	return ret
}

func (p *PrintableType) MarshalYAML() (interface{}, error) {
	ret := make(map[string]any)
	for _, f := range p.SimpleFields {
		ret[f.Name] = f.Value
	}

	for _, f := range p.ComplexField {
		ret[f.Name] = f.ToSliceOfMap()
	}
	return yaml.Marshal(ret)
}

func (l *ListType) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(l.ToSliceOfMap())
}

func (p *PrintableType) PrintTable(out io.Writer) {
	w := out
	if _, ok := out.(*tabwriter.Writer); !ok {
		w = tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
		defer w.(*tabwriter.Writer).Flush()
	}

	for _, f := range p.SimpleFields {
		fmt.Fprintf(w, "%s:\t%v\n", f.Name, f.Value)
	}

	for _, f := range p.ComplexField {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s:\n", f.Name)
		fmt.Fprintf(w, "\t%s\n", strings.Join(UpperCase(f.Headers), "\t"))
		for _, line := range f.Values {
			for _, v := range line {
				fmt.Fprintf(w, "\t%v", v)
			}
			fmt.Fprintln(w)
		}
	}
}
