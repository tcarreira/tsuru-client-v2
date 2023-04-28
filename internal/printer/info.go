// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import "io"

func PrintInfo(out io.Writer, format OutputType, data any, opts *TableViewOptions) (err error) {
	switch format {
	case JSON:
		return PrintJSON(out, data)
	case PrettyJSON:
		return PrintPrettyJSON(out, data)
	case YAML:
		return PrintYAML(out, data)
	case Table:
		if pData, ok := data.(PrintableType); ok {
			pData.PrintTable(out)
			return nil
		}
		return PrintTable(out, data, opts)
	default:
		if pData, ok := data.(PrintableType); ok {
			pData.PrintTable(out)
			return nil
		}
		return PrintTable(out, data, opts)
	}
}
