// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"golang.org/x/net/websocket"
)

func TestAppShellInfo(t *testing.T) {
	var stdout bytes.Buffer
	appShellCmd := newAppShellCmd()
	appShellCmd.SetOutput(&stdout)
	err := appShellCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestAppShellIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "shell" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand list not registered in appCmd")
}

func TestAppShellRunWithApp(t *testing.T) {
	stdout := bytes.Buffer{}
	expected := "hello my friend\nglad to see you here\n"
	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))

		fmt.Fprint(ws, expected)
		ws.Close()
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appShellCmd := newAppShellCmd()
	appShellCmd.Flags().Parse([]string{"--app", "myapp"})
	err = appShellCmdRun(appShellCmd, []string{}, apiClient, &stdout, stdin)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestShellToContainerWithUnit(t *testing.T) {
	stdout := bytes.Buffer{}
	expected := "hello my friend\nglad to see you here\n"
	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))
		assert.Equal(t, req.URL.Query().Get("unit"), "containerid")

		fmt.Fprint(ws, expected)
		ws.Close()
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appShellCmd := newAppShellCmd()
	appShellCmd.Flags().Parse([]string{"--app", "myapp"})
	err := appShellCmdRun(appShellCmd, []string{"containerid"}, apiClient, &stdout, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}
