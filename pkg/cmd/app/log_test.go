// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
)

func TestAppLogIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
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
	var stdout bytes.Buffer
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
	apiClient := api.APIClientWithConfig(
		&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()},
		&api.APIClientOpts{LocalTZ: time.UTC},
	)

	appLogCmd := newAppLogCmd()
	appLogCmd.Flags().Parse([]string{"--app", "appName"})
	err = appLogCmdRun(appLogCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}
