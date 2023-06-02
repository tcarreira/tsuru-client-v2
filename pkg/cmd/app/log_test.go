// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
	"github.com/tsuru/tsuru-client/pkg/printer"
)

func TestAppLogIsRegistered(t *testing.T) {
	appCmd := NewAppCmd(tsuructx.TsuruContextWithConfig(nil))
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "log" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand log not registered in appCmd")
}

func TestV1FormatterUsesCurrentTimeZone(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	t2 := t1.Add(2 * time.Hour)
	logs := []log{
		{Date: t1, Message: "Something happened", Source: "tsuru"},
		{Date: t2, Message: "Something happened again", Source: "tsuru"},
	}
	data, err := json.Marshal(logs)
	assert.NoError(t, err)
	var writer bytes.Buffer

	logFmt := logFormatter{localTZ: time.UTC}
	err = logFmt.Format(&writer, json.NewDecoder(bytes.NewReader(data)))
	assert.NoError(t, err)

	cfy := printer.Colorify{}
	expected := cfy.Colorfy(t1.Format(tLogFmt)+" [tsuru]:", "blue", "", "") + " Something happened\n"
	expected += cfy.Colorfy(t2.Format(tLogFmt)+" [tsuru]:", "blue", "", "") + " Something happened again\n"
	assert.Equal(t, expected, writer.String())
}

func TestV1AppLog(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	t2 := t1.Add(2 * time.Hour)
	logs := []log{
		{Date: t1, Message: "creating app lost", Source: "tsuru"},
		{Date: t2, Message: "app lost successfully created", Source: "app", Unit: "abcdef"},
	}

	result, err := json.Marshal(logs)
	assert.NoError(t, err)

	cfy := printer.Colorify{}
	expected := cfy.Colorfy(t1.Format(tLogFmt)+" [tsuru]:", "blue", "", "") + " creating app lost\n"
	expected += cfy.Colorfy(t2.Format(tLogFmt)+" [app][abcdef]:", "blue", "", "") + " app lost successfully created\n"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(result)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"--app", "appName"})
	err = appLogCmdRun(tsuruCtx, appLogCmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogWithUnparsableData(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	logs := []log{
		{Date: t1, Message: "creating app lost", Source: "tsuru"},
	}

	result, err := json.Marshal(logs)
	assert.NoError(t, err)

	cfy := printer.Colorify{}
	expected := cfy.Colorfy(t1.Format(tLogFmt)+" [tsuru]:", "blue", "", "") + " creating app lost\n"
	expected += "Error: unable to parse json: invalid character 'u' looking for beginning of value: \"\\nunparseable data\""

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(result)+"\nunparseable data")
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"--app", "appName"})
	err = appLogCmdRun(tsuruCtx, appLogCmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogWithoutTheFlag(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	t2 := t1.Add(2 * time.Hour)
	logs := []log{
		{Date: t1, Message: "creating app lost", Source: "tsuru"},
		{Date: t2, Message: "app lost successfully created", Source: "app"},
	}

	result, err := json.Marshal(logs)
	assert.NoError(t, err)

	cfy := printer.Colorify{}
	expected := cfy.Colorfy(t1.Format(tLogFmt)+" [tsuru]:", "blue", "", "") + " creating app lost\n"
	expected += cfy.Colorfy(t2.Format(tLogFmt)+" [app]:", "blue", "", "") + " app lost successfully created\n"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "10", r.URL.Query().Get("lines"))
		w.Write(result)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"--app", "hitthelights"})
	err = appLogCmdRun(tsuruCtx, appLogCmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogShouldReturnNilIfHasNoContent(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(nil)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"--app", "appName"})
	err := appLogCmdRun(tsuruCtx, appLogCmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "", tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogInfo(t *testing.T) {
	var stdout strings.Builder
	appLogCmd := newAppLogCmd(tsuructx.TsuruContextWithConfig(nil))
	appLogCmd.SetOutput(&stdout)
	err := appLogCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestV1AppLogBySource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "mysource", r.URL.Query().Get("source"))
		w.Write(nil)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--source", "mysource"})
	err := appLogCmdRun(tsuruCtx, appLogCmd, []string{})
	assert.NoError(t, err)
}

func TestV1AppLogByUnit(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "api", r.URL.Query().Get("unit"))
		w.Write(nil)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--unit", "api"})
	err := appLogCmdRun(tsuruCtx, appLogCmd, []string{})
	assert.NoError(t, err)
}

func TestV1AppLogWithLines(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "12", r.URL.Query().Get("lines"))
		w.Write(nil)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--lines", "12"})
	err := appLogCmdRun(tsuruCtx, appLogCmd, []string{})
	assert.NoError(t, err)
}

func TestV1AppLogWithFollow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "12", r.URL.Query().Get("lines"))
		assert.Equal(t, "1", r.URL.Query().Get("follow"))
		w.Write(nil)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--lines", "12", "-f"})
	err := appLogCmdRun(tsuruCtx, appLogCmd, []string{})
	assert.NoError(t, err)
}

func TestV1AppLogWithNoDateAndNoSource(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	t2 := t1.Add(2 * time.Hour)
	logs := []log{
		{Date: t1, Message: "GET /", Source: "web"},
		{Date: t2, Message: "POST /", Source: "web"},
	}

	result, err := json.Marshal(logs)
	assert.NoError(t, err)

	expected := "GET /\nPOST /\n"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "12", r.URL.Query().Get("lines"))
		assert.Equal(t, "1", r.URL.Query().Get("follow"))
		w.Write(result)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--lines", "12", "-f", "--no-date", "--no-source"})
	err = appLogCmdRun(tsuruCtx, appLogCmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogWithNoSource(t *testing.T) {
	t1 := time.Now().In(time.UTC)
	t2 := t1.Add(2 * time.Hour)
	logs := []log{
		{Date: t1, Message: "GET /", Source: "web"},
		{Date: t2, Message: "POST /", Source: "web"},
	}

	result, err := json.Marshal(logs)
	assert.NoError(t, err)

	cfy := printer.Colorify{}
	expected := cfy.Colorfy(t1.Format(tLogFmt)+":", "blue", "", "") + " GET /\n"
	expected += cfy.Colorfy(t2.Format(tLogFmt)+":", "blue", "", "") + " POST /\n"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps/hitthelights/log"))
		assert.Equal(t, "12", r.URL.Query().Get("lines"))
		assert.Equal(t, "1", r.URL.Query().Get("follow"))
		w.Write(result)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appLogCmd := newAppLogCmd(tsuruCtx)
	appLogCmd.Flags().Parse([]string{"-a", "hitthelights", "--lines", "12", "-f", "--no-source"})
	err = appLogCmdRun(tsuruCtx, appLogCmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppLogFlags(t *testing.T) {
	appLogCmd := newAppLogCmd(tsuructx.TsuruContextWithConfig(nil))
	flagset := appLogCmd.Flags()
	assert.NotNil(t, flagset)

	for _, test := range []struct {
		long     string
		usage    string
		toParse  []string
		expected string
	}{
		{"app", "The name of the app (may be passed as argument)", []string{"-a", "myapp"}, "myapp"},
		{"app", "The name of the app (may be passed as argument)", []string{"--app", "myapp2"}, "myapp2"},
		{"unit", "The log from the given unit (may be passed as argument)", []string{"-u", "myunit"}, "myunit"},
		{"unit", "The log from the given unit (may be passed as argument)", []string{"--unit", "myunit2"}, "myunit2"},
		{"lines", "The number of log lines to display", []string{"-l", "25"}, "25"},
		{"lines", "The number of log lines to display", []string{"--lines", "45"}, "45"},
		{"source", "The log from the given source", []string{"-s", "mysource"}, "mysource"},
		{"source", "The log from the given source", []string{"--source", "mysource2"}, "mysource2"},
		{"follow", "Follow logs", []string{}, "false"},
		{"follow", "Follow logs", []string{"-f"}, "true"},
		{"follow", "Follow logs", []string{"--follow"}, "true"},
		{"no-date", "No date information", []string{}, "false"},
		{"no-date", "No date information", []string{"--no-date"}, "true"},
		{"no-source", "No source information", []string{}, "false"},
		{"no-source", "No source information", []string{"--no-source"}, "true"},
	} {
		err := flagset.Parse(test.toParse)
		assert.NoError(t, err)
		flag := flagset.Lookup(test.long)
		assert.Equal(t, test.long, flag.Name)
		assert.Equal(t, test.usage, flag.Usage)
		assert.Equal(t, test.expected, flag.Value.String())
	}
}
