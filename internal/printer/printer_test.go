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
		{"pretty-json", PrettyJSON},
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
		{"test", `"test"`},
		{[]byte("test"), `"dGVzdA=="`},
		{&bytes.Buffer{}, "{}"},
		{map[string]string{"key": "value"}, `{"key":"value"}`},
		{map[string]interface{}{"key": map[string]string{"subkey": "value"}}, `{"key":{"subkey":"value"}}`},
		{struct {
			Key   string
			Value int
		}{"mykey", 42}, `{"Key":"mykey","Value":42}`},
		{tsuru.Plan{Name: "myplan", Memory: 1024, Cpumilli: 1000}, `{"name":"myplan","memory":1024,"cpumilli":1000,"override":{}}`},
		{(*tsuru.Plan)(nil), "null"},
	} {
		w := &bytes.Buffer{}
		err := PrintJSON(w, test.data)
		assert.NoError(t, err)
		assert.EqualValues(t, test.expected+"\n", w.String())
	}

	// Empty result (passing nil)
	for _, test := range []struct {
		data interface{}
	}{
		{nil},
	} {
		w := &bytes.Buffer{}
		err := PrintJSON(w, test.data)
		assert.NoError(t, err)
		assert.EqualValues(t, "", w.String())
	}

	// Error cases
	for _, test := range []struct {
		data interface{}
	}{
		{func() {}},
		{make(chan bool)},
	} {
		w := &bytes.Buffer{}
		err := PrintJSON(w, test.data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error converting to json")
	}
}

func TestPrintPrettyJSON(t *testing.T) {
	for _, test := range []struct {
		data     interface{}
		expected string
	}{
		{"test", `"test"`},
		{[]byte("test"), `"dGVzdA=="`},
		{&bytes.Buffer{}, "{}"},
		{map[string]string{"key": "value"}, "{\n  \"key\": \"value\"\n}"},
		{map[string]interface{}{"key": map[string]string{"subkey": "value"}}, `{
  "key": {
    "subkey": "value"
  }
}`},
		{struct {
			Key   string
			Value int
		}{"mykey", 42}, `{
  "Key": "mykey",
  "Value": 42
}`},
		{tsuru.Plan{Name: "myplan", Memory: 1024, Cpumilli: 1000}, `{
  "name": "myplan",
  "memory": 1024,
  "cpumilli": 1000,
  "override": {}
}`},
		{(*tsuru.Plan)(nil), "null"},
	} {
		w := &bytes.Buffer{}
		err := PrintPrettyJSON(w, test.data)
		assert.NoError(t, err)
		assert.EqualValues(t, test.expected+"\n", w.String())
	}

	// Empty result (passing nil)
	for _, test := range []struct {
		data interface{}
	}{
		{nil},
	} {
		w := &bytes.Buffer{}
		err := PrintPrettyJSON(w, test.data)
		assert.NoError(t, err)
		assert.EqualValues(t, "", w.String())
	}

	// Error cases
	for _, test := range []struct {
		data interface{}
	}{
		{func() {}},
		{make(chan bool)},
	} {
		w := &bytes.Buffer{}
		err := PrintPrettyJSON(w, test.data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error converting to json")
	}
}
