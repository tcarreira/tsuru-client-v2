// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"

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
	Verbosity          int
	VerboseOutput      io.Writer
	InsecureSkipVerify bool
}

type tsuruClientHTTPTransport struct {
	t             http.RoundTripper
	cfg           *tsuru.Configuration
	apiClientOpts *APIClientOpts
}

func (t *tsuruClientHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.cfg.DefaultHeader {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", t.cfg.UserAgent)
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "tsuru-client")
	}
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "bearer sometoken")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	if t.apiClientOpts != nil && t.apiClientOpts.InsecureSkipVerify {
		t.t.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	req.Close = true

	req.Header.Set("X-Tsuru-Verbosity", "0")
	// Verbosity level=1: log request
	if t.apiClientOpts != nil && t.apiClientOpts.Verbosity >= 1 {
		req.Header.Set("X-Tsuru-Verbosity", strconv.Itoa(t.apiClientOpts.Verbosity))
		fmt.Fprintf(t.apiClientOpts.VerboseOutput, "*************************** <Request uri=%q> **********************************\n", req.URL.RequestURI())
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprint(t.apiClientOpts.VerboseOutput, string(requestDump))
		if requestDump[len(requestDump)-1] != '\n' {
			fmt.Fprintln(t.apiClientOpts.VerboseOutput)
		}
		fmt.Fprintf(t.apiClientOpts.VerboseOutput, "*************************** </Request uri=%q> **********************************\n", req.URL.RequestURI())
	}

	response, err := t.t.RoundTrip(req)

	// Verbosity level=2: log response
	if t.apiClientOpts != nil && t.apiClientOpts.Verbosity >= 2 && response != nil {
		fmt.Fprintf(t.apiClientOpts.VerboseOutput, "*************************** <Response uri=%q> **********************************\n", req.URL.RequestURI())
		responseDump, errDump := httputil.DumpResponse(response, true)
		if errDump != nil {
			return nil, errDump
		}
		fmt.Fprint(t.apiClientOpts.VerboseOutput, string(responseDump))
		if responseDump[len(responseDump)-1] != '\n' {
			fmt.Fprintln(t.apiClientOpts.VerboseOutput)
		}
		fmt.Fprintf(t.apiClientOpts.VerboseOutput, "*************************** </Response uri=%q> **********************************\n", req.URL.RequestURI())
	}

	return response, err
}

func httpTransportWrapper(cfg *tsuru.Configuration, apiClientOpts *APIClientOpts, roundTripper http.RoundTripper) *tsuruClientHTTPTransport {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &tsuruClientHTTPTransport{
		t:             roundTripper,
		cfg:           cfg,
		apiClientOpts: apiClientOpts,
	}
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

	cfg.HTTPClient.Transport = httpTransportWrapper(cfg, opts, cfg.HTTPClient.Transport)

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
