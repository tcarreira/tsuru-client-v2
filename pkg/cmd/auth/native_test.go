// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func TestNativeLogin(t *testing.T) {
	viper.Set("token", "") // concurrent with TestLoginCmdRunErr
	result := `{"token": "sometoken", "is_admin": true}`
	expected := "Email: Password: \nSuccessfully logged in!\n"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/users/foo@foo.com/tokens"))
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "chico", r.FormValue("password"))
		fmt.Fprint(w, result)
	}))

	tsuruCtx := tsuructx.TsuruContextWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)
	tsuruCtx.AuthScheme = "native"
	tsuruCtx.Stdin = &tsuructx.FakeStdin{Reader: strings.NewReader("foo@foo.com\nchico\n")}

	cmd := NewLoginCmd()
	err := loginCmdRun(cmd, []string{}, tsuruCtx)
	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestNativeLoginWithoutEmailFromArg(t *testing.T) {
	viper.Set("token", "") // concurrent with TestLoginCmdRunErr
	result := `{"token": "sometoken", "is_admin": true}`
	expected := "Password: \nSuccessfully logged in!\n"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/users/foo@foo.com/tokens"))
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "chico", r.FormValue("password"))
		fmt.Fprint(w, result)
	}))

	tsuruCtx := tsuructx.TsuruContextWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)
	tsuruCtx.AuthScheme = "native"
	tsuruCtx.Stdin = &tsuructx.FakeStdin{Reader: strings.NewReader("chico\n")}

	cmd := NewLoginCmd()
	err := loginCmdRun(cmd, []string{"foo@foo.com"}, tsuruCtx)
	assert.NoError(t, err)
	assert.Equal(t, expected, tsuruCtx.Stdout.(*strings.Builder).String())
}
