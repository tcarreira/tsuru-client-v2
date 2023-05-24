// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuructx

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/exec"
)

var tsuruContextSingleton *TsuruContext

type TsuruContext struct {
	TsuruContextOpts

	// Client is the tsuru client implementated by go-tsuruclient
	Client *tsuru.APIClient
	// Config is the tsuru client configuration
	Config *tsuru.Configuration
	// RawHTTPClient is the raw http client for REST calls
	RawHTTPClient *http.Client
}

type TsuruContextOpts struct {
	// Verbosity is the verbosity level for tsuru client. Should be 1 ou 2
	Verbosity int
	// InsecureSkipVerify will skip TLS verification (not applied to websockets)
	InsecureSkipVerify bool
	// LocalTZ is the local timezone
	LocalTZ *time.Location
	// AuthScheme is the protocol used for tsuru login. Overriden with TSURU_AUTH_SCHEME
	AuthScheme string
	// Executor is an instance of an interface for exec.Command()
	Executor exec.Executor
	//Fs is the filesystem used by the client
	Fs afero.Fs

	Stdout io.Writer
	Stderr io.Writer
	Stdin  DescriptorReader
}

type DescriptorReader interface {
	Read(p []byte) (n int, err error)
	Fd() uintptr
}

func GetTsuruContextSingleton() *TsuruContext {
	if tsuruContextSingleton == nil {
		SetupTsuruContextSingleton(nil, nil)
	}
	return tsuruContextSingleton
}

// SetupTsuruContextSingleton configures the tsuruContext to be returned by GetTsuruContextSingleton().
func SetupTsuruContextSingleton(cfg *tsuru.Configuration, opts *TsuruContextOpts) {
	tsuruContextSingleton = TsuruContextWithConfig(cfg, opts)
}

func DefaultTestingTsuruContextOptions() *TsuruContextOpts {
	return &TsuruContextOpts{
		LocalTZ:  time.UTC,
		Fs:       afero.NewMemMapFs(),
		Executor: &exec.FakeExec{},

		Stdout: &strings.Builder{},
		Stderr: &strings.Builder{},
		Stdin:  &FakeStdin{strings.NewReader("")},
	}
}

// TsuruContextWithConfig returns a new TsuruContext with the given configuration.
func TsuruContextWithConfig(cfg *tsuru.Configuration, opts *TsuruContextOpts) *TsuruContext {
	if cfg == nil {
		cfg = tsuru.NewConfiguration()
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	cfg.DefaultHeader = tsuruDefaultHeadersFromConfig(cfg)

	if opts == nil {
		// defaults for testing
		opts = DefaultTestingTsuruContextOptions()
	}

	tsuruCtx := &TsuruContext{
		Client:           tsuru.NewAPIClient(cfg),
		RawHTTPClient:    cfg.HTTPClient,
		Config:           cfg,
		TsuruContextOpts: *opts,
	}

	transportOpts := &ClientHTTPTransportOpts{
		InsecureSkipVerify: &tsuruCtx.InsecureSkipVerify,
		Verbosity:          &tsuruCtx.Verbosity,
		VerboseOutput:      &tsuruCtx.Stdout,
	}
	tsuruCtx.Config.HTTPClient.Transport = httpTransportWrapper(cfg, transportOpts, cfg.HTTPClient.Transport)

	return tsuruCtx
}

var _ DescriptorReader = &FakeStdin{}

type FakeStdin struct {
	Reader io.Reader
}

func (f *FakeStdin) Read(p []byte) (n int, err error) {
	return f.Reader.Read(p)
}
func (f *FakeStdin) Fd() uintptr {
	return 0
}
