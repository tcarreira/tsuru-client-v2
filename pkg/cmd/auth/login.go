// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func NewLoginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login [EMAIL]",
		Short: "initiates a new tsuru session for a user",
		Long: `Initiates a new tsuru session for a user. If using tsuru native authentication
scheme, it will ask for the email and the password and check if the user is
successfully authenticated. If using OAuth, it will open a web browser for the
user to complete the login.

After that, the token generated by the tsuru server will be stored in
[[${HOME}/.tsuru/token]].

All tsuru actions require the user to be authenticated (except [[tsuru login]]
and [[tsuru version]]).
`,
		Example: `$ tsuru login
$ tsuru login example@tsuru.local`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return loginCmdRun(cmd, args, tsuructx.GetTsuruContextSingleton())
		},
		Args: cobra.RangeArgs(0, 1),
	}

	return loginCmd
}

func loginCmdRun(cmd *cobra.Command, args []string, tsuruCtx *tsuructx.TsuruContext) error {
	return fmt.Errorf("not implemented yet")
}
