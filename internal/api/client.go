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

var apiClientSingleton *APIClient

type APIClient struct {
	Client        *tsuru.APIClient
	RawHTTPClient *http.Client
	Config        *tsuru.Configuration
	Opts          *APIClientOpts
}

type APIClientOpts struct {
	Verbosity int
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
	if viper.GetString("insecure-skip-verify") == "true" {
		t.t.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return t
}

// APIClientSingleton returns the APIClient singleton configured with SetupAPIClientSingleton().
func APIClientSingleton() *APIClient {
	if apiClientSingleton == nil {
		SetupAPIClientSingleton(nil, nil)
	}
	return apiClientSingleton
}

// SetupAPIClientSingleton configures the tsuru APIClient to be returned by APIClientSingleton().
func SetupAPIClientSingleton(cfg *tsuru.Configuration, opts *APIClientOpts) {
	apiClientSingleton = APIClientWithConfig(cfg, opts)
}

// APIClientWithConfig returns a new APIClient with the given configuration.
func APIClientWithConfig(cfg *tsuru.Configuration, opts *APIClientOpts) *APIClient {
	if cfg == nil {
		cfg = tsuru.NewConfiguration()
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.HTTPClient.Transport == nil {
		cfg.HTTPClient.Transport = newTsuruClientHTTPTransport(cfg)
	}
	return &APIClient{
		Client:        tsuru.NewAPIClient(cfg),
		RawHTTPClient: cfg.HTTPClient,
		Config:        cfg,
		Opts:          opts,
	}
}

// NewRequest creates a new http.Request with the correct base path.
func (a *APIClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, a.Config.BasePath) {
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		if !regexp.MustCompile(`^/[0-9]+\.[0-9]+/`).MatchString(url) {
			url = "/1.0" + url
		}
		url = strings.TrimRight(a.Config.BasePath, "/") + url
	}
	return http.NewRequest(method, url, body)
}
