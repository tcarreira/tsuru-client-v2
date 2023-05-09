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

func TestV1AppCreateEmptyPlatform(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps"))
		assert.Equal(t, "ble", r.FormValue("name"))
		assert.Equal(t, "", r.FormValue("platform"))
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
	err := appCreateRun(appCreateCmd, []string{"ble"}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())
}

func TestV1AppCreateTeamOwner(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps"))
		assert.Equal(t, "ble", r.FormValue("name"))
		assert.Equal(t, "django", r.FormValue("platform"))
		assert.Equal(t, "myteam", r.FormValue("teamOwner"))
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
	appCreateCmd.LocalFlags().Parse([]string{"-t", "myteam"})
	var stdout bytes.Buffer
	err := appCreateRun(appCreateCmd, []string{"ble", "django"}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())
}

func TestV1AppCreatePlan(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps"))
		assert.Equal(t, "ble", r.FormValue("name"))
		assert.Equal(t, "django", r.FormValue("platform"))
		assert.Equal(t, "", r.FormValue("teamOwner"))
		assert.Equal(t, "myplan", r.FormValue("plan"))
		assert.Equal(t, "", r.FormValue("pool"))
		assert.Equal(t, "", r.FormValue("description"))
		assert.Equal(t, "", r.FormValue("router"))
		r.ParseForm()
		assert.Nil(t, r.Form["tag"])

		fmt.Fprintln(w, `{"status":"success", "repository_url":"git@tsuru.plataformas.glb.com:ble.git"}`)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appCreateCmd := newAppCreateCmd()
	appCreateCmd.LocalFlags().Parse([]string{"-p", "myplan"})
	var stdout bytes.Buffer
	err := appCreateRun(appCreateCmd, []string{"ble", "django"}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())
}

func TestV1AppCreatePool(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		assert.True(t, strings.HasSuffix(r.URL.Path, "/apps"))
		assert.Equal(t, "ble", r.FormValue("name"))
		assert.Equal(t, "django", r.FormValue("platform"))
		assert.Equal(t, "", r.FormValue("teamOwner"))
		assert.Equal(t, "", r.FormValue("plan"))
		assert.Equal(t, "mypool", r.FormValue("pool"))
		assert.Equal(t, "", r.FormValue("description"))
		assert.Equal(t, "", r.FormValue("router"))
		r.ParseForm()
		assert.Nil(t, r.Form["tag"])

		fmt.Fprintln(w, `{"status":"success", "repository_url":"git@tsuru.plataformas.glb.com:ble.git"}`)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appCreateCmd := newAppCreateCmd()
	appCreateCmd.LocalFlags().Parse([]string{"-o", "mypool"})
	var stdout bytes.Buffer
	err := appCreateRun(appCreateCmd, []string{"ble", "django"}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())
}

func TestV1AppCreateWithTags(t *testing.T) {
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
		if assert.Equal(t, 2, len(r.Form["tag"])) {
			assert.Equal(t, "tag1", r.Form["tag"][0])
			assert.Equal(t, "tag2", r.Form["tag"][1])
		}

		fmt.Fprintln(w, `{"status":"success", "repository_url":"git@tsuru.plataformas.glb.com:ble.git"}`)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appCreateCmd := newAppCreateCmd()
	appCreateCmd.LocalFlags().Parse([]string{"--tag", "tag1", "--tag", "tag2"})
	var stdout bytes.Buffer
	err := appCreateRun(appCreateCmd, []string{"ble", "django"}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), stdout.String())
}
