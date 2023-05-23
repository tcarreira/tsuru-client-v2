// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuructx

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

type ClientHTTPTransportOpts struct {
	InsecureSkipVerify bool
	Verbosity          int
	VerboseOutput      *io.Writer
}

type TsuruClientHTTPTransport struct {
	ClientHTTPTransportOpts

	t   http.RoundTripper
	cfg *tsuru.Configuration
}

func (t *TsuruClientHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.cfg.DefaultHeader {
		req.Header.Set(k, v)
	}

	if t.InsecureSkipVerify {
		t.t.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	req.Close = true

	req.Header.Set("X-Tsuru-Verbosity", "0")
	// Verbosity level=1: log request
	if t.Verbosity >= 1 {
		req.Header.Set("X-Tsuru-Verbosity", strconv.Itoa(t.Verbosity))
		fmt.Fprintf(*t.VerboseOutput, "*************************** <Request uri=%q> **********************************\n", req.URL.RequestURI())
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprint(*t.VerboseOutput, string(requestDump))
		if requestDump[len(requestDump)-1] != '\n' {
			fmt.Fprintln(*t.VerboseOutput)
		}
		fmt.Fprintf(*t.VerboseOutput, "*************************** </Request uri=%q> **********************************\n", req.URL.RequestURI())
	}

	response, err := t.t.RoundTrip(req)

	// Verbosity level=2: log response
	if t.Verbosity >= 2 && response != nil {
		fmt.Fprintf(*t.VerboseOutput, "*************************** <Response uri=%q> **********************************\n", req.URL.RequestURI())
		responseDump, errDump := httputil.DumpResponse(response, true)
		if errDump != nil {
			return nil, errDump
		}
		fmt.Fprint(*t.VerboseOutput, string(responseDump))
		if responseDump[len(responseDump)-1] != '\n' {
			fmt.Fprintln(*t.VerboseOutput)
		}
		fmt.Fprintf(*t.VerboseOutput, "*************************** </Response uri=%q> **********************************\n", req.URL.RequestURI())
	}

	return response, err
}

func httpTransportWrapper(cfg *tsuru.Configuration, opts *ClientHTTPTransportOpts, roundTripper http.RoundTripper) *TsuruClientHTTPTransport {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &TsuruClientHTTPTransport{
		t:                       roundTripper,
		cfg:                     cfg,
		ClientHTTPTransportOpts: *opts,
	}
}

func tsuruDefaultHeadersFromConfig(cfg *tsuru.Configuration) map[string]string {
	result := map[string]string{}
	for k, v := range cfg.DefaultHeader {
		result[k] = v
	}

	result["User-Agent"] = cfg.UserAgent
	if result["User-Agent"] == "" {
		result["User-Agent"] = "tsuru-client"
	}
	if result["Authorization"] == "" {
		result["Authorization"] = "bearer sometoken"
	}
	if result["Accept"] == "" {
		result["Accept"] = "application/json"
	}
	return result
}

// NewRequest creates a new http.Request with the correct base path.
func (tc *TsuruContext) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, tc.Config.BasePath) {
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		if !regexp.MustCompile(`^/[0-9]+\.[0-9]+/`).MatchString(url) {
			url = "/1.0" + url
		}
		url = strings.TrimRight(tc.Config.BasePath, "/") + url
	}
	return http.NewRequest(method, url, body)
}

func (tc *TsuruContext) DefaultHeaders() http.Header {
	headers := make(http.Header)
	for k, v := range tc.Config.DefaultHeader {
		headers.Add(k, v)
	}
	return headers
}
