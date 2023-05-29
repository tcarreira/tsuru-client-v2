// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func newAppDeployCmd() *cobra.Command {
	appDeployCmd := &cobra.Command{
		Use:   "deploy APP [file-or-dir ...]",
		Short: "deploy the source code and/or configurations to the application on Tsuru",
		Long: `Deploy the source code and/or configurations to the application on Tsuru.
Files specified in the ".tsuruignore" file are skipped - similar to ".gitignore". It also honors ".dockerignore" file if deploying with container file (--dockerfile).
`,
		Example: `To deploy using app's platform build process (just sending source code and/or configurations):
  Uploading all files within the current directory
    $ tsuru app deploy -a <APP> .

  Uploading all files within a specific directory
    $ tsuru app deploy -a <APP> mysite/

  Uploading specific files
    $ tsuru app deploy -a <APP> ./myfile.jar ./Procfile

  Uploading specific files (ignoring their base directories)
    $ tsuru app deploy -a <APP> --files-only ./my-code/main.go ./tsuru_stuff/Procfile

To deploy using a container image:
    $ tsuru app deploy -a <APP> --image registry.example.com/my-company/app:v42

To deploy using container file ("docker build" mode):
  Sending the the current directory as container build context - uses Dockerfile file as container image instructions:
    $ tsuru app deploy -a <APP> --dockerfile .

  Sending a specific container file and specific directory as container build context:
    $ tsuru app deploy -a <APP> --dockerfile ./Dockerfile.other ./other/
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appDeployCmdRun(cmd, args, tsuructx.GetTsuruContextSingleton())
		},
		Args: cobra.MinimumNArgs(0),
	}

	appDeployCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appDeployCmd.Flags().StringP("image", "i", "", "The image to deploy in app")
	appDeployCmd.Flags().StringP("message", "m", "", "A message describing this deploy")
	appDeployCmd.Flags().BoolP("files-only", "f", false, "Enables single file deployment into the root of the app's tree")
	appDeployCmd.Flags().String("dockerfile", "", "Container file")
	return appDeployCmd
}

func appDeployCmdRun(cmd *cobra.Command, args []string, tsuruCtx *tsuructx.TsuruContext) error {
	return nil
}
