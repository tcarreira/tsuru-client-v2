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
	"github.com/spf13/viper"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/exec"
)

var tsuruContextSingleton *TsuruContext

type TsuruContext struct {
	TsuruContextOpts
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
	// Fs is the filesystem used by the client
	Fs afero.Fs
	// Viper is an instance of the viper.Viper configuration
	Viper *viper.Viper

	UserAgent string
	TargetURL string
	Token     string

	Stdout io.Writer
	Stderr io.Writer
	Stdin  DescriptorReader
}

// Config is the tsuru client configuration
func (c *TsuruContext) Config() *tsuru.Configuration {
	cfg := tsuru.NewConfiguration()
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	cfg.BasePath = c.TargetURL
	cfg.UserAgent = c.UserAgent
	cfg.HTTPClient.Transport = c.httpTransportWrapper(cfg.HTTPClient.Transport)
	return cfg
}

// Client is the tsuru client implementated by go-tsuruclient
func (c *TsuruContext) Client() *tsuru.APIClient {
	return tsuru.NewAPIClient(c.Config())
}

// RawHTTPClient is the raw http client for REST calls
func (c *TsuruContext) RawHTTPClient() *http.Client {
	return c.Config().HTTPClient
}

type DescriptorReader interface {
	Read(p []byte) (n int, err error)
	Fd() uintptr
}

func GetTsuruContextSingleton() *TsuruContext {
	if tsuruContextSingleton == nil {
		SetupTsuruContextSingleton(nil)
	}
	return tsuruContextSingleton
}

// SetupTsuruContextSingleton configures the tsuruContext to be returned by GetTsuruContextSingleton().
func SetupTsuruContextSingleton(opts *TsuruContextOpts) {
	tsuruContextSingleton = TsuruContextWithConfig(opts)
}

func DefaultTestingTsuruContextOptions() *TsuruContextOpts {
	return &TsuruContextOpts{
		Verbosity:          0,
		InsecureSkipVerify: false,
		LocalTZ:            time.UTC,
		AuthScheme:         "",
		Executor:           &exec.FakeExec{},
		Fs:                 afero.NewMemMapFs(),

		UserAgent: "tsuru-client:testing",
		TargetURL: "",
		Token:     "sometoken",

		Stdout: &strings.Builder{},
		Stderr: &strings.Builder{},
		Stdin:  &FakeStdin{strings.NewReader("")},
	}
}

// TsuruContextWithConfig returns a new TsuruContext with the given configuration.
func TsuruContextWithConfig(opts *TsuruContextOpts) *TsuruContext {
	if opts == nil {
		// defaults for testing
		opts = DefaultTestingTsuruContextOptions()
	}

	tsuruCtx := &TsuruContext{
		TsuruContextOpts: *opts,
	}

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
