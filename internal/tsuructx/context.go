package tsuructx

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
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

	Stdout io.Writer
	Stderr io.Writer
	Stdin  descriptorReader
}

type TsuruContextOpts struct {
	// Verbosity is the verbosity level for tsuru client. Should be 1 ou 2
	Verbosity int
	// InsecureSkipVerify will skip TLS verification (not applied to websockets)
	InsecureSkipVerify bool
	// LocalTZ is the local timezone
	LocalTZ *time.Location
}

type descriptorReader interface {
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
		opts = &TsuruContextOpts{
			LocalTZ: time.Local,
		}
	}

	tsuruCtx := &TsuruContext{
		Client:           tsuru.NewAPIClient(cfg),
		RawHTTPClient:    cfg.HTTPClient,
		Config:           cfg,
		TsuruContextOpts: *opts,

		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	transportOpts := &ClientHTTPTransportOpts{
		InsecureSkipVerify: opts.InsecureSkipVerify,
		Verbosity:          opts.Verbosity,
		VerboseOutput:      tsuruCtx.Stdout,
	}
	tsuruCtx.Config.HTTPClient.Transport = httpTransportWrapper(cfg, transportOpts, cfg.HTTPClient.Transport)

	return tsuruCtx
}
