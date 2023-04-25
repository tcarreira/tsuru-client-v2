// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

var tsuruClient *tsuru.APIClient

// Client returns a tsuru client configured with SetupTsuruClient() or with the default configuration.
func Client() *tsuru.APIClient {
	if tsuruClient == nil {
		tsuruClient = tsuru.NewAPIClient(&tsuru.Configuration{})
	}
	return tsuruClient
}

// SetupTsuruClient configures the tsuru APIClient to be returned by Client().
func SetupTsuruClient(cfg *tsuru.Configuration) {
	tsuruClient = tsuru.NewAPIClient(cfg)
}
