// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
)

func TestAppInfo(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"Ip":"10.10.10.10","ID":"app1/0","Status":"started","Address":{"Host": "10.8.7.6:3333"}}, {"Ip":"9.9.9.9","ID":"app1/1","Status":"started","Address":{"Host": "10.8.7.6:3323"}}, {"Ip":"","ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+----------+------+
| Name   | Status  | Host     | Port |
+--------+---------+----------+------+
| app1/2 | pending |          |      |
| app1/0 | started | 10.8.7.6 | 3333 |
| app1/1 | started | 10.8.7.6 | 3323 |
+--------+---------+----------+------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoSimplified(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","pool": "dev-a", "provisioner": "kubernetes", "cluster": "mycluster", "teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"Ip":"10.10.10.10","ID":"app1/0","Status":"started","ProcessName": "web","Address":{"Host": "10.8.7.6:3333"}, "ready": true, "routable": true}, {"Ip":"9.9.9.9","ID":"app1/1","Status":"started","ProcessName": "web","Address":{"Host": "10.8.7.6:3323"}, "ready": true, "routable": true}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb", "plan":{"name": "test",  "memory": 536870912, "cpumilli": 100, "default": false}}`
	expected := `Application: app1
Created by: myapp_owner
Platform: php
Plan: test
Pool: dev-a (kubernetes | cluster: mycluster)
Router: planb
Teams: myteam (owner), tsuruteam, crane
Cluster External Addresses: myapp.tsuru.io
Units: 2
+---------+-------+----------+---------------+------------+
| Process | Ready | Restarts | Avg CPU (abs) | Avg Memory |
+---------+-------+----------+---------------+------------+
| web     | 2/2   | 0        | 0%            | 0Mi        |
+---------+-------+----------+---------------+------------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1", "-s"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)
	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoKubernetes(t *testing.T) {
	var stdout bytes.Buffer
	t0 := time.Now().UTC().Format(time.RFC3339)
	t1 := time.Now().Add(time.Hour * -1).UTC().Format(time.RFC3339)
	t2 := time.Now().Add(time.Hour * -1 * 24 * 30).UTC().Format(time.RFC3339)

	result := fmt.Sprintf(`{
		"name":"app1",
		"teamowner":"myteam",
		"cname":[""],"ip":"myapp.tsuru.io",
		"provisioner": "kubernetes",
		"platform":"php",
		"repository":"git@git.com:php.git",
		"state":"dead",
		"cluster": "kube-cluster-dev",
		"pool": "dev-a",
		"units":[
			{"Ip":"10.10.10.10","ID":"app1/0","Status":"started","Address":{"Host": "10.8.7.6:3333"}, "ready": true, "restarts": 10, "createdAt": "%s"},
			{"Ip":"9.9.9.9","ID":"app1/1","Status":"started","Address":{"Host": "10.8.7.6:3323"}, "ready": true, "restarts": 0, "createdAt": "%s"},
			{"Ip":"","ID":"app1/2","Status":"pending", "ready": false, "createdAt": "%s"}
		],
		"unitsMetrics": [
			{"ID": "app1/0", "CPU": "900m", "Memory": "2000000Ki"},
			{"ID": "app1/1", "CPU": "800m", "Memory": "3000000Ki"},
			{"ID": "app1/2", "CPU": "80m", "Memory": "300Ki"}
		],
		"teams": ["tsuruteam","crane"],
		"owner": "myapp_owner",
		"deploys": 7,
		"router": "planb"
	}`, t0, t1, t2)
	expected := `Application: app1
Platform: php
Provisioner: kubernetes
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Cluster: kube-cluster-dev
Pool: dev-a
Quota: 0/0 units

Units: 3
+--------+----------+---------+----------+-----+-----+--------+
| Name   | Host     | Status  | Restarts | Age | CPU | Memory |
+--------+----------+---------+----------+-----+-----+--------+
| app1/2 |          | pending |          | 30d | 8%  | 0Mi    |
| app1/0 | 10.8.7.6 | ready   | 10       | 0s  | 90% | 1953Mi |
| app1/1 | 10.8.7.6 | ready   | 0        | 60m | 80% | 2929Mi |
+--------+----------+---------+----------+-----+-----+--------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoMultipleAddresses(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"Ip":"10.10.10.10","ID":"app1/0","Status":"started","Address":{"Host": "10.8.7.6:3333"},"Addresses":[{"Host": "10.8.7.6:3333"}, {"Host": "10.8.7.6:4444"}]}, {"Ip":"9.9.9.9","ID":"app1/1","Status":"started","Address":{"Host": "10.8.7.6:3323"}}, {"Ip":"","ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+----------+------------+
| Name   | Status  | Host     | Port       |
+--------+---------+----------+------------+
| app1/2 | pending |          |            |
| app1/0 | started | 10.8.7.6 | 3333, 4444 |
| app1/1 | started | 10.8.7.6 | 3323       |
+--------+---------+----------+------------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoMultipleRouters(t *testing.T) {
	var stdout bytes.Buffer
	result := `
{
	"name": "app1",
	"teamowner": "myteam",
	"cname": [
		"cname1"
	],
	"ip": "myapp.tsuru.io",
	"platform": "php",
	"repository": "git@git.com:php.git",
	"state": "dead",
	"units": [
		{
			"Ip": "10.10.10.10",
			"ID": "app1/0",
			"Status": "started",
			"Address": {
				"Host": "10.8.7.6:3333"
			}
		},
		{
			"Ip": "9.9.9.9",
			"ID": "app1/1",
			"Status": "started",
			"Address": {
				"Host": "10.8.7.6:3323"
			}
		},
		{
			"Ip": "",
			"ID": "app1/2",
			"Status": "pending"
		}
	],
	"teams": [
		"tsuruteam",
		"crane"
	],
	"owner": "myapp_owner",
	"deploys": 7,
	"router": "planb",
	"routers": [
		{"name": "r1", "type": "r", "opts": {"a": "b", "x": "y"}, "address": "addr1"},
		{"name": "r2", "addresses": ["addr2", "addr9"], "status": "ready"},
		{"name": "r3", "type": "r3", "address": "addr3", "status": "not ready", "status-detail": "something happening"}
	]
}`
	expected := `Application: app1
Platform: php
Teams: myteam (owner), tsuruteam, crane
External Addresses: cname1 (cname), addr1, addr2, addr9, addr3
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+----------+------+
| Name   | Status  | Host     | Port |
+--------+---------+----------+------+
| app1/2 | pending |          |      |
| app1/0 | started | 10.8.7.6 | 3333 |
| app1/1 | started | 10.8.7.6 | 3323 |
+--------+---------+----------+------+

Routers:
+------+------+-----------+--------------------------------+
| Name | Opts | Addresses | Status                         |
+------+------+-----------+--------------------------------+
| r1   | a: b | addr1     |                                |
|      | x: y |           |                                |
+------+------+-----------+--------------------------------+
| r2   |      | addr2     | ready                          |
|      |      | addr9     |                                |
+------+------+-----------+--------------------------------+
| r3   |      | addr3     | not ready: something happening |
+------+------+-----------+--------------------------------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithDescription(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "description": "My app", "router": "planb"}`
	expected := `Application: app1
Description: My app
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithTags(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"Ip":"10.10.10.10","ID":"app1/0","Status":"started"}, {"Ip":"9.9.9.9","ID":"app1/1","Status":"started"}, {"Ip":"","ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "tags": ["tag 1", "tag 2", "tag 3"], "router": "planb"}`
	expected := `Application: app1
Tags: tag 1, tag 2, tag 3
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+-------------+------+
| Name   | Status  | Host        | Port |
+--------+---------+-------------+------+
| app1/2 | pending |             |      |
| app1/0 | started | 10.10.10.10 |      |
| app1/1 | started | 9.9.9.9     |      |
+--------+---------+-------------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithRouterOpts(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "routeropts": {"opt1": "val1", "opt2": "val2"}, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb (opt1=val1, opt2=val2)
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithQuota(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb", "quota": {"inUse": 3, "limit": 40}}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 3/40 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoLock(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"Ip":"","ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "lock": {"locked": true, "owner": "admin@example.com", "reason": "DELETE /apps/rbsample/units", "acquiredate": "2012-04-01T10:32:00Z"}, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Lock:
 Acquired in: 2012-04-01 10:32:00 +0000 UTC
 Owner: admin@example.com
 Running: DELETE /apps/rbsample/units
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoManyProcesses(t *testing.T) {
	var stdout bytes.Buffer
	result := `{
  "name": "app1",
  "teamowner": "myteam",
  "cname": [
    ""
  ],
  "ip": "myapp.tsuru.io",
  "platform": "php",
  "repository": "git@git.com:php.git",
  "state": "dead",
  "units": [
    {
      "Ip": "10.10.10.10",
      "ID": "app1/0",
      "Status": "started",
      "ProcessName": "web"
    },
    {
      "Ip": "9.9.9.9",
      "ID": "app1/1",
      "Status": "started",
      "ProcessName": "worker"
    },
    {
      "Ip": "",
      "ID": "app1/2",
      "Status": "pending",
      "ProcessName": "worker"
    }
  ],
  "teams": [
    "tsuruteam",
    "crane"
  ],
  "owner": "myapp_owner",
  "deploys": 7,
  "router": "planb"
}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units [process web]: 1
+--------+---------+-------------+------+
| Name   | Status  | Host        | Port |
+--------+---------+-------------+------+
| app1/0 | started | 10.10.10.10 |      |
+--------+---------+-------------+------+

Units [process worker]: 2
+--------+---------+---------+------+
| Name   | Status  | Host    | Port |
+--------+---------+---------+------+
| app1/2 | pending |         |      |
| app1/1 | started | 9.9.9.9 |      |
+--------+---------+---------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoManyVersions(t *testing.T) {
	var stdout bytes.Buffer
	result := `{
  "name": "app1",
  "teamowner": "myteam",
  "cname": [
    ""
  ],
  "ip": "myapp.tsuru.io",
  "platform": "php",
  "repository": "git@git.com:php.git",
  "state": "dead",
  "units": [
    {
      "ID": "app1/0",
      "Status": "started",
	  "ProcessName": "web",
	  "Version": 1,
	  "Routable": false
    },
    {
      "ID": "app1/1",
      "Status": "started",
	  "ProcessName": "worker",
	  "Version": 1,
	  "Routable": false
    },
    {
      "ID": "app1/2",
      "Status": "pending",
	  "ProcessName": "worker",
	  "Version": 1,
	  "Routable": false
	},
	{
      "ID": "app1/3",
      "Status": "started",
	  "ProcessName": "web",
	  "Version": 2,
	  "Routable": true
    },
    {
      "ID": "app1/4",
      "Status": "started",
	  "ProcessName": "worker",
	  "Version": 2,
	  "Routable": true
    }
  ],
  "teams": [
    "tsuruteam",
    "crane"
  ],
  "owner": "myapp_owner",
  "deploys": 7,
  "router": "planb"
}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units [process web] [version 1]: 1
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
+--------+---------+------+------+

Units [process worker] [version 1]: 2
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

Units [process web] [version 2] [routable]: 1
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/3 | started |      |      |
+--------+---------+------+------+

Units [process worker] [version 2] [routable]: 1
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/4 | started |      |      |
+--------+---------+------+------+

`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithAutoScale(t *testing.T) {
	var stdout bytes.Buffer
	result := `{
  "name": "app1",
  "teamowner": "myteam",
  "cname": [
    ""
  ],
  "ip": "myapp.tsuru.io",
  "platform": "php",
  "repository": "git@git.com:php.git",
  "state": "dead",
  "units": [
    {
      "ID": "app1/0",
      "Status": "started",
      "ProcessName": "web"
    },
    {
      "ID": "app1/1",
      "Status": "started",
      "ProcessName": "worker"
    }
  ],
  "teams": [
    "tsuruteam",
    "crane"
  ],
  "owner": "myapp_owner",
  "deploys": 7,
  "router": "planb",
  "autoscale": [
    {
      "process":"web",
      "minUnits":1,
      "maxUnits":10,
      "averageCPU":"500m",
      "version":10
    },
    {
      "process":"worker",
      "minUnits":2,
      "maxUnits":5,
      "averageCPU":"2",
      "version":10
    }
  ]
}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units [process web]: 1
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
+--------+---------+------+------+

Units [process worker]: 1
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/1 | started |      |      |
+--------+---------+------+------+

Auto Scale:
+--------------+-----+-----+------------+
| Process      | Min | Max | Target CPU |
+--------------+-----+-----+------------+
| web (v10)    | 1   | 10  | 50%        |
| worker (v10) | 2   | 5   | 200%       |
+--------------+-----+-----+------------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoNoUnits(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","ip":"app1.tsuru.io","teamowner":"myteam","platform":"php","repository":"git@git.com:php.git","state":"dead","units":[],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: app1.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoEmptyUnit(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"x","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"Name":"","Status":""}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: x (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"--app", "app1"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"secret","teamowner":"myteam","ip":"secret.tsuru.io","platform":"ruby","repository":"git@git.com:php.git","state":"dead","units":[{"Ip":"","ID":"secret/0","Status":"started"}, {"Ip":"","ID":"secret/1","Status":"pending"}],"Teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb", "quota": {"inUse": 0, "limit": -1}}`
	expected := `Application: secret
Platform: ruby
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: secret.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/unlimited

Units: 2
+----------+---------+------+------+
| Name     | Status  | Host | Port |
+----------+---------+------+------+
| secret/0 | started |      |      |
| secret/1 | pending |      |      |
+----------+---------+------+------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/apps/secret") && r.Method == "GET" {
			fmt.Fprintln(w, result)
		}
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"-a", "secret"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoCName(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","ip":"myapp.tsuru.io","cname":["yourapp.tsuru.io"],"platform":"php","repository":"git@git.com:php.git","state":"dead","units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"Ip":"","ID":"app1/2","Status":"pending"}],"Teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: yourapp.tsuru.io (cname), myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"-a", "secret"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithServices(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead","units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"Ip":"","ID":"app1/2","Status":"pending"}],"Teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb", "serviceInstanceBinds": [{"service": "redisapi", "instance": "myredisapi"}]}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

Service instances: 1
+----------+-----------------+
| Service  | Instance (Plan) |
+----------+-----------------+
| redisapi | myredisapi      |
+----------+-----------------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"-a", "secret"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithServicesTwoService(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead","units":[{"Ip":"10.10.10.10","ID":"app1/0","Status":"started"}, {"Ip":"9.9.9.9","ID":"app1/1","Status":"started"}, {"Ip":"","ID":"app1/2","Status":"pending"}],"Teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "router": "planb", "serviceInstanceBinds": [{"service": "redisapi", "instance": "myredisapi"}, {"service": "mongodb", "instance": "mongoapi"}]}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+-------------+------+
| Name   | Status  | Host        | Port |
+--------+---------+-------------+------+
| app1/2 | pending |             |      |
| app1/0 | started | 10.10.10.10 |      |
| app1/1 | started | 9.9.9.9     |      |
+--------+---------+-------------+------+

Service instances: 2
+----------+-----------------+
| Service  | Instance (Plan) |
+----------+-----------------+
| mongodb  | mongoapi        |
| redisapi | myredisapi      |
+----------+-----------------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"-a", "secret"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}

func TestAppInfoWithPlan(t *testing.T) {
	var stdout bytes.Buffer
	result := `{"name":"app1","teamowner":"myteam","cname":[""],"ip":"myapp.tsuru.io","platform":"php","repository":"git@git.com:php.git","state":"dead", "units":[{"ID":"app1/0","Status":"started"}, {"ID":"app1/1","Status":"started"}, {"ID":"app1/2","Status":"pending"}],"teams":["tsuruteam","crane"], "owner": "myapp_owner", "deploys": 7, "plan":{"name": "test",  "memory": 536870912, "cpumilli": 100, "default": false}, "router": "planb"}`
	expected := `Application: app1
Platform: php
Router: planb
Teams: myteam (owner), tsuruteam, crane
External Addresses: myapp.tsuru.io
Created by: myapp_owner
Deploys: 7
Pool:
Quota: 0/0 units

Units: 3
+--------+---------+------+------+
| Name   | Status  | Host | Port |
+--------+---------+------+------+
| app1/0 | started |      |      |
| app1/1 | started |      |      |
| app1/2 | pending |      |      |
+--------+---------+------+------+

App Plan:
+------+-----+--------+
| Name | CPU | Memory |
+------+-----+--------+
| test | 10% | 512Mi  |
+------+-----+--------+

`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, result)
	}))
	apiClient := api.APIClientWithConfig(&tsuru.Configuration{BasePath: mockServer.URL, HTTPClient: mockServer.Client()})

	appInfoCmd := newAppInfoCmd()
	appInfoCmd.Flags().Parse([]string{"-a", "secret"})
	err := printAppInfo(appInfoCmd, []string{}, apiClient, &stdout)

	assert.NoError(t, err)
	assert.Equal(t, expected, stdout.String())
}
