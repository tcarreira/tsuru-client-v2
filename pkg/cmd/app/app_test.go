// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
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
