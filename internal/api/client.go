// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"crypto/tls"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

var (
	tsuruClient   *tsuru.APIClient
	tsuruCfg      *tsuru.Configuration
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

// NewRequest creates a new http.Request with the correct base path.
func NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, tsuruCfg.BasePath) {
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		if !regexp.MustCompile(`^/[0-9]+\.[0-9]+/`).MatchString(url) {
			url = "/1.0" + url
		}
		url = strings.TrimRight(tsuruCfg.BasePath, "/") + url
	}
	return http.NewRequest(method, url, body)
}

func newTsuruClientHTTPTransport(cfg *tsuru.Configuration) *tsuruClientHTTPTransport {
	t := &tsuruClientHTTPTransport{
		t:   http.DefaultTransport,
		cfg: cfg,
	}
	if viper.GetString("insecure-skip-verify") == "true" {
		t.t.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return t
}

// SetupTsuruClient configures the tsuru APIClient to be returned by Client().
func SetupTsuruClient(cfg *tsuru.Configuration) {
	if cfg == nil {
		cfg = &tsuru.Configuration{}
	}
	tsuruCfg = cfg
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	rawHTTPClient = cfg.HTTPClient
	if rawHTTPClient.Transport == nil {
		rawHTTPClient.Transport = newTsuruClientHTTPTransport(cfg)
	}
	tsuruClient = tsuru.NewAPIClient(cfg)
}
