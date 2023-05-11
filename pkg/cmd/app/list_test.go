// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
)

func TestV1AppList(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[{"ID":"app1/0","Status":"started"}]}]`
	expected := `+-------------+-----------+-------------+
| Application | Units     | Address     |
+-------------+-----------+-------------+
| app1        | 1 started | 10.10.10.10 |
+-------------+-----------+-------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListDisplayAppsInAlphabeticalOrder(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.11","name":"sapp","units":[{"ID":"sapp1/0","Status":"started"}]},{"ip":"10.10.10.10","name":"app1","units":[{"ID":"app1/0","Status":"started"}]}]`
	expected := `+-------------+-----------+-------------+
| Application | Units     | Address     |
+-------------+-----------+-------------+
| app1        | 1 started | 10.10.10.10 |
+-------------+-----------+-------------+
| sapp        | 1 started | 10.10.10.11 |
+-------------+-----------+-------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListUnitIsntAvailable(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[{"ID":"app1/0","Status":"pending"}]}]`
	expected := `+-------------+-----------+-------------+
| Application | Units     | Address     |
+-------------+-----------+-------------+
| app1        | 1 pending | 10.10.10.10 |
+-------------+-----------+-------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListErrorFetchingUnits(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[],"Error": "timeout"}]`
	expected := `+-------------+----------------------+-------------+
| Application | Units                | Address     |
+-------------+----------------------+-------------+
| app1        | error fetching units | 10.10.10.10 |
+-------------+----------------------+-------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListErrorFetchingUnitsVerbose(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[],"Error": "timeout"}]`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(
		&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()},
		&api.APIClientOpts{Verbosity: 1, VerboseOutput: &stdout},
	)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	expected := "*************************** <Request uri=\"/1.0/apps\"> **********************************\n" +
		"GET /1.0/apps HTTP/1.1\r\n" +
		"Host: " + strings.Split(mockServer.URL, "://")[1] + "\r\n" +
		"Accept: application/json\r\n" +
		"Authorization: bearer sometoken\r\n" +
		"User-Agent: tsuru-client\r\n" +
		"X-Tsuru-Verbosity: 1\r\n" +
		"\r\n" +
		"*************************** </Request uri=\"/1.0/apps\"> **********************************\n" +
		"+-------------+-------------------------------+-------------+\n" +
		"| Application | Units                         | Address     |\n" +
		"+-------------+-------------------------------+-------------+\n" +
		"| app1        | error fetching units: timeout | 10.10.10.10 |\n" +
		"+-------------+-------------------------------+-------------+\n"

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListUnitWithoutID(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[{"ID":"","Status":"pending"}, {"ID":"unit2","Status":"stopped"}]}]`
	expected := `+-------------+-----------+-------------+
| Application | Units     | Address     |
+-------------+-----------+-------------+
| app1        | 1 stopped | 10.10.10.10 |
+-------------+-----------+-------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppListCName(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","cname":["app1.tsuru.io"],"name":"app1","units":[{"ID":"app1/0","Status":"started"}]}]`
	expected := `+-------------+-----------+-----------------------+
| Application | Units     | Address               |
+-------------+-----------+-----------------------+
| app1        | 1 started | app1.tsuru.io (cname) |
|             |           | 10.10.10.10           |
+-------------+-----------+-----------------------+
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListFiltering(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","cname":["app1.tsuru.io"],"name":"app1","units":[{"ID":"app1/0","Status":"started"}]}]`
	expected := `+-------------+-----------+-----------------------+
| Application | Units     | Address               |
+-------------+-----------+-----------------------+
| app1        | 1 started | app1.tsuru.io (cname) |
|             |           | 10.10.10.10           |
+-------------+-----------+-----------------------+
`
	expectedQueryString := url.Values(map[string][]string{
		"platform":  {"python"},
		"locked":    {"true"},
		"owner":     {"glenda@tsuru.io"},
		"teamOwner": {"tsuru"},
		"name":      {"myapp"},
		"pool":      {"pool"},
		"status":    {"started"},
		"tag":       {"tag a", "tag b"},
	})
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, expectedQueryString, r.URL.Query())
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	appListCmd.Flags().Parse([]string{"-p", "python", "--locked", "--user", "glenda@tsuru.io", "-t", "tsuru", "--name", "myapp", "--pool", "pool", "--status", "started", "--tag", "tag a", "--tag", "tag b"})

	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListFilteringMe(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","cname":["app1.tsuru.io"],"name":"app1","units":[{"ID":"app1/0","Status":"started"}]}]`
	expected := `+-------------+-----------+-----------------------+
| Application | Units     | Address               |
+-------------+-----------+-----------------------+
| app1        | 1 started | app1.tsuru.io (cname) |
|             |           | 10.10.10.10           |
+-------------+-----------+-----------------------+
`
	httpServerState := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch httpServerState {
		case 0:
			fmt.Fprintln(w, `{"Email":"gopher@tsuru.io","Teams":[]}`)
		case 1:
			assert.EqualValues(t, url.Values(map[string][]string{"owner": {"gopher@tsuru.io"}}), r.URL.Query())
			fmt.Fprintln(w, result)
		default:
			assert.Fail(t, "unexpected httpServerState: %d", httpServerState)
		}
		httpServerState++
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	appListCmd.Flags().Parse([]string{"-u", "me"})

	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListSortByCountAndStatus(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","cname":["app1.tsuru.io"],"name":"app1","units":[{"ID":"app1/0","Status":"starting"},{"ID":"app1/1","Status":"stopped"},{"ID":"app1/2","Status":"asleep"},{"ID":"app1/3","Status":"started"},{"ID":"app1/4","Status":"started"},{"ID":"app1/5","Status":"stopped"}]}]`
	expected := `+-------------+------------+-----------------------+
| Application | Units      | Address               |
+-------------+------------+-----------------------+
| app1        | 2 started  | app1.tsuru.io (cname) |
|             | 2 stopped  | 10.10.10.10           |
|             | 1 asleep   |                       |
|             | 1 starting |                       |
+-------------+------------+-----------------------+
`
	httpServerState := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch httpServerState {
		case 0:
			fmt.Fprintln(w, `{"Email":"gopher@tsuru.io","Teams":[]}`)
		case 1:
			assert.EqualValues(t, url.Values(map[string][]string{"owner": {"gopher@tsuru.io"}}), r.URL.Query())
			fmt.Fprintln(w, result)
		default:
			assert.Fail(t, "unexpected httpServerState: %d", httpServerState)
		}
		httpServerState++
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	appListCmd.Flags().Parse([]string{"-u", "me"})

	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListWithFlagQ(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"ip":"10.10.10.10","name":"app1","units":[{"ID":"app1/0","Status":"started"}]},{"ip":"10.10.10.11","name":"app2","units":[{"ID":"app2/0","Status":"started"}]},{"ip":"10.10.10.12","cname":["app3.tsuru.io"],"name":"app3","units":[{"ID":"app3/0","Status":"started"}]}]`
	expected := `app1
app2
app3
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, url.Values(map[string][]string{"simplified": {"true"}}), r.URL.Query())
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	appListCmd.Flags().Parse([]string{"-q"})

	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestV1AppListWithFlags(t *testing.T) {
	var stdout bytes.Buffer
	result := `[{"name":"app1","platform":"python","pool":"pool2"},{"name":"app2","platform":"python","pool":"pool2"},{"name":"app3","platform":"go","pool":"pool1"}]`
	expected := `app1
app2
app3
`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, url.Values(map[string][]string{
			"platform":   {"python"},
			"simplified": {"true"},
		}), r.URL.Query())
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()}, nil)

	appListCmd := newAppListCmd()
	appListCmd.Flags().Parse([]string{"-p", "python", "-q"})

	err := appListCmdRun(appListCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppListIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "list" {
			found = true
		}
	}
	assert.True(t, found, "subcommand list not registered in appCmd")
}
