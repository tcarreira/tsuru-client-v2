package printer

import (
	"io"
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
