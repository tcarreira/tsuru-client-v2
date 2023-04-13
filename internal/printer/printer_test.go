package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
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

func TestPrintJSON(t *testing.T) {
	for _, test := range []struct {
		data     interface{}
		expected string
	}{
		{nil, "null"},
		{"test", `"test"`},
		{[]byte("test"), `"dGVzdA=="`},
		{map[string]string{"key": "value"}, `{"key":"value"}`},
		{map[string]interface{}{"key": map[string]string{"subkey": "value"}}, `{"key":{"subkey":"value"}}`},
		{struct {
			Key   string
			Value int
		}{"mykey", 42}, `{"Key":"mykey","Value":42}`},
		{tsuru.Plan{Name: "myplan", Memory: 1024, Cpumilli: 1000}, `{"name":"myplan","memory":1024,"cpumilli":1000,"override":{}}`},
	} {
		w := &bytes.Buffer{}
		err := PrintJSON(w, test.data)
		assert.NoError(t, err)
		assert.EqualValues(t, test.expected+"\n", w.String())
	}
}
