package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
)

const expectedFmt = `App %q has been created!
Use app info to check the status of the app and its units.
`

func TestV1AppCreate(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps"))
		assert.Equal(t, "ble", r.FormValue("name"))
		assert.Equal(t, "django", r.FormValue("platform"))
		assert.Equal(t, "", r.FormValue("teamOwner"))
		assert.Equal(t, "", r.FormValue("plan"))
		assert.Equal(t, "", r.FormValue("pool"))
		assert.Equal(t, "", r.FormValue("description"))
		assert.Equal(t, "", r.FormValue("router"))
		r.ParseForm()
		assert.Nil(t, r.Form["tag"])

		fmt.Fprintln(w, `{"status":"success", "repository_url":"git@tsuru.plataformas.glb.com:ble.git"}`)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appCreateCmd := newAppCreateCmd()
	var stdout bytes.Buffer
	err := appCreateRun(appCreateCmd, []string{"ble", "django"}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())

}
