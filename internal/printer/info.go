// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import (
	"io"
	"text/tabwriter"
)

func PrintInfo(out io.Writer, format OutputType, data any) (err error) {
	switch format {
	case JSON:
		return PrintJSON(out, data)
	case PrettyJSON:
		return PrintPrettyJSON(out, data)
	case YAML:
		return PrintYAML(out, data)
	case Table:
		return PrintTable(out, data)
	default:
		return PrintTable(out, data)
	}
}

func PrintList(out io.Writer, format OutputType, data any) (err error) {
	switch format {
	case JSON:
		return PrintJSON(out, data)
	case PrettyJSON:
		return PrintPrettyJSON(out, data)
	case YAML:
		return PrintYAML(out, data)
	case Table:
		return PrintSubTable(out, data)
	default:
		return PrintSubTable(out, data)
	}
}

func PrintSubTable(out io.Writer, data any) error {
	w := tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
	defer w.Flush()
	return printSubTable(w, data)
}
