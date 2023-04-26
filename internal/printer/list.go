package printer

import "io"

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
