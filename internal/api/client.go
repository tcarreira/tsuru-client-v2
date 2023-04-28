// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

var (
	tsuruClient   *tsuru.APIClient
	rawHTTPClient *http.Client
)

// Client returns a tsuru client configured with SetupTsuruClient() or with the default configuration.
func Client() *tsuru.APIClient {
	if tsuruClient == nil {
		SetupTsuruClient(&tsuru.Configuration{})
	}
	return tsuruClient
}

func RawHTTPClient() *http.Client {
	return rawHTTPClient
}

type tsuruClientHTTPTransport struct {
	t   http.RoundTripper
	cfg *tsuru.Configuration
}

func (t *tsuruClientHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.cfg.UserAgent)
	for k, v := range t.cfg.DefaultHeader {
		req.Header.Set(k, v)
	}
	return t.t.RoundTrip(req)
}

func newTsuruClientHTTPTransport(cfg *tsuru.Configuration) *tsuruClientHTTPTransport {
	t := &tsuruClientHTTPTransport{
		t:   http.DefaultTransport,
		cfg: cfg,
	}
	return t
}

// SetupTsuruClient configures the tsuru APIClient to be returned by Client().
func SetupTsuruClient(cfg *tsuru.Configuration) {
	if cfg == nil {
		cfg = &tsuru.Configuration{}
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	rawHTTPClient = cfg.HTTPClient
	if rawHTTPClient.Transport == nil {
		rawHTTPClient.Transport = newTsuruClientHTTPTransport(cfg)
	}
	tsuruClient = tsuru.NewAPIClient(cfg)
}
