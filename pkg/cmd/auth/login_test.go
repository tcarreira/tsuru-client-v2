// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func TestNewLoginCmd(t *testing.T) {
	assert.NotNil(t, NewLoginCmd())
}

func TestLoginCmdRunErr(t *testing.T) {
	viper.Set("token", "xxx")
	err := loginCmdRun(nil, nil, nil)
	assert.EqualError(t, err, "this command can't run with $TSURU_TOKEN environment variable set. Did you forget to unset?")
}

func TestGetAuthScheme(t *testing.T) {
	result := "{\"name\":\"oauth\",\"data\":{\"authorizeUrl\":\"https://auth.tsuru.local/authorize?client_id=xpto\\u0026redirect_uri=__redirect_url__\\u0026response_type=code\\u0026scope=user\",\"port\":\"12345\"}}\n"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))

	tsuruCtx := tsuructx.TsuruContextWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	authScheme, err := getAuthScheme(tsuruCtx)
	assert.NoError(t, err)
	assert.Equal(t, "oauth", authScheme.Name)
	assert.Equal(t, "12345", authScheme.Data["port"])
}
