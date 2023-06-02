// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"-t", "myteam"})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"-p", "myplan"})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"-o", "mypool"})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppCreateRouterOpts(t *testing.T) {
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
		assert.Equal(t, "1", r.FormValue("routeropts.a"))
		assert.Equal(t, "2", r.FormValue("routeropts.b"))

		fmt.Fprintln(w, `{"status":"success"}`)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"--router-opts", "a=1", "--router-opts", "b=2"})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppCreateNoRepository(t *testing.T) {
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

		fmt.Fprintln(w, `{"status":"success"}`)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppCreateWithInvalidFramework(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "")
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{})
	assert.Error(t, err)
	assert.Equal(t, "", tsuruCtx.Stdout.(*strings.Builder).String())
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
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"--tag", "tag1", "--tag", "tag2"})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppCreateWithEmptyTag(t *testing.T) {
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
		// if assert.Equal(t, 1, len(r.Form["tag"])) {
		// 	assert.Equal(t, "", r.Form["tag"][0])
		// }
		assert.Equal(t, 0, len(r.Form["tag"])) // XXX: breaking test from V1

		fmt.Fprintln(w, `{"status":"success", "repository_url":"git@tsuru.plataformas.glb.com:ble.git"}`)
	}))
	tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
	tsuruCtx.SetTargetURL(mockServer.URL)

	appCreateCmd := newAppCreateCmd(tsuruCtx)
	appCreateCmd.Flags().Parse([]string{"--tag", ""})
	err := appCreateRun(tsuruCtx, appCreateCmd, []string{"ble", "django"})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedFmt, "ble"), tsuruCtx.Stdout.(*strings.Builder).String())
}

func TestV1AppCreateInfo(t *testing.T) {
	var stdout strings.Builder
	appCreateCmd := newAppCreateCmd(tsuructx.TsuruContextWithConfig(nil))
	appCreateCmd.SetOutput(&stdout)
	err := appCreateCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestV1AppCreateFlags(t *testing.T) {
	appCreateCmd := newAppCreateCmd(tsuructx.TsuruContextWithConfig(nil))
	flagset := appCreateCmd.Flags()
	assert.NotNil(t, flagset)

	for _, test := range []struct {
		short   string
		long    string
		usage   string
		toParse []string
	}{
		{"-p", "plan", "the plan used to create the app", []string{"-p", "myplan"}},
		{"-p", "plan", "the plan used to create the app", []string{"--plan", "myplan2"}},
		{"-t", "team", "team owning the app", []string{"-t", "myteam"}},
		{"-t", "team", "team owning the app", []string{"--team", "myteam2"}},
		{"-r", "router", "the router used by the app", []string{"-r", "myrouter"}},
		{"-r", "router", "the router used by the app", []string{"--router", "myrouter2"}},
	} {
		err := flagset.Parse(test.toParse)
		assert.NoError(t, err)
		flag := flagset.Lookup(test.long)
		assert.Equal(t, test.long, flag.Name)
		assert.Equal(t, test.usage, flag.Usage)
		assert.Equal(t, test.toParse[1], flag.Value.String())
	}

	err := flagset.Parse([]string{"--tag", "tag1", "--tag", "tag2"})
	assert.NoError(t, err)
	tagFlag := flagset.Lookup("tag")
	assert.Equal(t, "tag", tagFlag.Name)
	assert.Equal(t, "app tags", tagFlag.Usage)
	assert.Equal(t, `[tag1,tag2]`, tagFlag.Value.String())

	err = flagset.Parse([]string{"-g", "tag3", "-g", "tag4"})
	assert.NoError(t, err)
	tagFlag = flagset.Lookup("tag")
	assert.Equal(t, "tag", tagFlag.Name)
	assert.Equal(t, "app tags", tagFlag.Usage)
	assert.Equal(t, `[tag1,tag2,tag3,tag4]`, tagFlag.Value.String())

	err = flagset.Parse([]string{"--router-opts", "opt1=val1", "--router-opts", "opt2=val2"})
	assert.NoError(t, err)
	optsFlag := flagset.Lookup("router-opts")
	assert.Equal(t, "router-opts", optsFlag.Name)
	assert.Equal(t, "router options", optsFlag.Usage)
	assert.Equal(t, `[opt1=val1,opt2=val2]`, optsFlag.Value.String())
}

func TestAppCreateIsRegistered(t *testing.T) {
	appCmd := NewAppCmd(tsuructx.TsuruContextWithConfig(nil))
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "create" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand create not registered in appCmd")
}
