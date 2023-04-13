package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatAs(t *testing.T) {
	for _, test := range []struct {
		s        string
		expected OutputType
	}{
		{"json", JSON},
		{"yaml", YAML},
		{"table", Table},
		{"invalid", Table},
	} {
		got := FormatAs(test.s)
		assert.Equal(t, test.expected, got)
	}
}
