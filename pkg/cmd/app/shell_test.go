// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
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

func TestV1AppShellRunWithApp(t *testing.T) {
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
	err := appShellCmdRun(appShellCmd, []string{}, apiClient, &stdout, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppShellWithUnit(t *testing.T) {
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

func TestAppShellWithUnitAppFromArgs(t *testing.T) {
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
	err := appShellCmdRun(appShellCmd, []string{"myapp", "containerid"}, apiClient, &stdout, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppShellCmdConnectionRefused(t *testing.T) {
	stdout := bytes.Buffer{}
	mockServer := httptest.NewServer(nil)
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)
	mockServer.Close()

	appShellCmd := newAppShellCmd()
	err := appShellCmdRun(appShellCmd, []string{"myapp"}, apiClient, &stdout, nil)
	assert.ErrorContains(t, err, "connection refused")
}

// newFileWithContent may be used to create a mock for Stdin.
func newFileWithContent(content []byte) (stdin *os.File, deferFn func(), err error) {
	deferFn = func() {}
	stdin, err = os.CreateTemp("", "")
	if err != nil {
		return
	}
	deferFn = func() {
		os.Remove(stdin.Name())
	}
	stdin.Write(content)
	stdin.Seek(0, 0)
	return
}

func TestAppShellSendStdin(t *testing.T) {
	stdout := bytes.Buffer{}
	expected := "hello my friend\n"
	stdin, deferFn, err := newFileWithContent([]byte(expected))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		deferFn()
	}()

	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))
		assert.Equal(t, req.URL.Query().Get("unit"), "containerid")

		fmt.Fprint(ws, "from websocket server\n")

		var buf = make([]byte, 1024)
		n, err := ws.Read(buf)
		assert.NoError(t, err, io.EOF)
		assert.Equal(t, expected, string(buf[:n]))

		n, err = ws.Read(buf)
		assert.ErrorIs(t, err, io.EOF)
		assert.Equal(t, 0, n)

		ws.Close()
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appShellCmd := newAppShellCmd()
	err = appShellCmdRun(appShellCmd, []string{"myapp", "containerid"}, apiClient, &stdout, stdin)
	assert.NoError(t, err)
	assert.Equal(t, "from websocket server\n", stdout.String())
}
