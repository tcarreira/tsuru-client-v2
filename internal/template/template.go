package template

import (
	"strings"
	"text/template"

	"github.com/tsuru/tsuru-client/internal/parser"
)

func DefaultTemplateFuncs() template.FuncMap {
	ret := make(template.FuncMap)
	ret["Age"] = SimpleAge
	ret["Join"] = Join
	return ret
}

// SimpleAge parses a time.RFC3339 string (2006-01-02T15:04:05Z07:00)
// and returns the duration since then in a human readable format (no seconds).
func SimpleAge(timeStr string) string {
	return parser.DurationFromTimeWithoutSeconds(timeStr, "")
}

func Join(args []string) string {
	return strings.Join(args, ", ")
}
