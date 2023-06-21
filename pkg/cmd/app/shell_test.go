// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/v2/internal/tsuructx"
	"golang.org/x/net/websocket"
)

func TestAppShellInfo(t *testing.T) {
	stdout := strings.Builder{}
	appShellCmd := newAppShellCmd(tsuructx.TsuruContextWithConfig(nil))
	appShellCmd.SetOutput(&stdout)
	err := appShellCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestAppShellIsRegistered(t *testing.T) {
	appCmd := NewAppCmd(tsuructx.TsuruContextWithConfig(nil))
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
	assert.True(t, found, "subcommand shell not registered in appCmd")
}

func TestV1AppShellRunWithApp(t *testing.T) {
	expected := "hello my friend\nglad to see you here"
	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))

		fmt.Fprint(ws, expected)
		ws.Close()
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appShellCmd := newAppShellCmd(tsuruCtx)
	appShellCmd.Flags().Parse([]string{"--app", "myapp"})
	err := appShellCmdRun(tsuruCtx, appShellCmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, expected+"\n", tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppShellWithUnit(t *testing.T) {
	expected := "hello my friend\nglad to see you here"
	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))
		assert.Equal(t, req.URL.Query().Get("unit"), "containerid")

		fmt.Fprint(ws, expected)
		ws.Close()
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	tsuruCtx.Stdin = nil

	appShellCmd := newAppShellCmd(tsuruCtx)
	appShellCmd.Flags().Parse([]string{"--app", "myapp"})
	err := appShellCmdRun(tsuruCtx, appShellCmd, []string{"containerid"})
	assert.NoError(t, err)
	assert.Equal(t, expected+"\n", tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestAppShellWithUnitAppFromArgs(t *testing.T) {
	expected := "hello my friend\nglad to see you here"
	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))
		assert.Equal(t, req.URL.Query().Get("unit"), "containerid")

		fmt.Fprint(ws, expected)
		ws.Close()
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	tsuruCtx.Stdin = nil

	appShellCmd := newAppShellCmd(tsuruCtx)
	err := appShellCmdRun(tsuruCtx, appShellCmd, []string{"myapp", "containerid"})
	assert.NoError(t, err)
	assert.Equal(t, expected+"\n", tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppShellCmdConnectionRefused(t *testing.T) {
	mockServer := httptest.NewServer(nil)
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	tsuruCtx.Stdin = nil
	mockServer.Close()

	appShellCmd := newAppShellCmd(tsuruCtx)
	err := appShellCmdRun(tsuruCtx, appShellCmd, []string{"myapp"})
	assert.ErrorContains(t, err, "refused") // windows: connectex: No connection could be made because the target machine actively refused it.
	// assert.ErrorContains(t, err, "connection refused") // unix: connect: connection refused
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
	strFromStdin := "hello my friend"
	strFromServer := "from websocket server"
	stdin, deferFn, err := newFileWithContent([]byte(strFromStdin))
	if err != nil {
		t.Fatal(err)
	}
	defer deferFn()

	mockServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		req := ws.Request()
		assert.NotNil(t, req)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/apps/myapp/shell"))
		assert.Equal(t, req.URL.Query().Get("unit"), "containerid")

		fmt.Fprint(ws, strFromServer)

		var buf = make([]byte, 1024)
		n, err := ws.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, strFromStdin, string(buf[:n]))

		ws.Close()
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	tsuruCtx.Stdin = stdin

	appShellCmd := newAppShellCmd(tsuruCtx)
	err = appShellCmdRun(tsuruCtx, appShellCmd, []string{"myapp", "containerid"})
	assert.NoError(t, err)
	assert.Equal(t, strFromServer+"\n", tsuruCtx.Stdout.(*strings.Builder).String())
}
