// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/internal/config"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func TestPort(t *testing.T) {
	assert.Equal(t, ":0", port(map[string]string{}))
	assert.Equal(t, ":4242", port(map[string]string{"port": "4242"}))
}

func TestCallbackHandler(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"token": "xpto"}`)
	}))
	defer mockServer.Close()

	redirectURL := "someurl"
	finish := make(chan bool, 1)
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.TargetURL = mockServer.URL

	callbackHandler := callback(tsuruCtx, redirectURL, finish)
	request, err := http.NewRequest("GET", "/", strings.NewReader(`{"code":"xpto"}`))
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	callbackHandler(recorder, request)

	assert.Equal(t, true, <-finish)
	assert.Equal(t, fmt.Sprintf(callbackPage, successMarkup), recorder.Body.String())
	file, err := tsuruCtx.Fs.Open(filepath.Join(config.ConfigPath, "token"))
	assert.NoError(t, err)
	data, err := io.ReadAll(file)
	assert.NoError(t, err)
	assert.Equal(t, "xpto", string(data))
}
