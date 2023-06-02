// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func newAppDeployCmd(tsuruCtx *tsuructx.TsuruContext) *cobra.Command {
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
			return appDeployCmdRun(tsuruCtx, cmd, args)
		},
		Args: cobra.MinimumNArgs(0),
	}

	appDeployCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appDeployCmd.Flags().StringP("image", "i", "", "The image to deploy in app")
	appDeployCmd.Flags().StringP("message", "m", "", "A message describing this deploy")
	appDeployCmd.Flags().BoolP("files-only", "f", false, "Enables single file deployment into the root of the app's tree")
	appDeployCmd.Flags().String("dockerfile", "", "Container file")
	appDeployCmd.Flags().Bool("new-version", false, "Creates a new version for the current deployment while preserving existing versions")
	appDeployCmd.Flags().Bool("override-old-versions", false, "Force replace all deployed versions by this new deploy")
	return appDeployCmd
}

func appDeployCmdRun(tsuruCtx *tsuructx.TsuruContext, cmd *cobra.Command, args []string) error {
	appName := cmd.Flag("app").Value.String()
	if appName == "" && len(args) > 0 {
		appName = args[0]
		args = args[1:]
	}

	if appName == "" {
		return fmt.Errorf("no app was provided. Please provide an app name")
	}

	if cmd.Flag("image").Value.String() == "" && cmd.Flag("dockerfile").Value.String() == "" && len(args) == 0 {
		return fmt.Errorf("you should provide at least one file, Docker image name or Dockerfile to deploy")
	}

	if cmd.Flag("image").Value.String() != "" && len(args) > 0 {
		return fmt.Errorf("you can't deploy files and docker image at the same time")
	}

	if cmd.Flag("image").Value.String() != "" && cmd.Flag("dockerfile").Value.String() != "" {
		return fmt.Errorf("you can't deploy container image and container file at same time")
	}

	cmd.SilenceUsage = true

	values := url.Values{}
	values.Set("origin", "app-deploy")
	if cmd.Flag("image").Value.String() != "" {
		values.Set("origin", "image")
	}
	if msg := cmd.Flag("message").Value.String(); msg != "" {
		values.Set("message", msg)
	}
	if newV := cmd.Flag("new-version").Value.String(); newV == "true" {
		values.Set("new-version", "true")
	}
	if overrideV := cmd.Flag("override-old-versions").Value.String(); overrideV == "true" {
		values.Set("override-versions", "true")
	}

	requestReader, _ := io.Pipe()
	request, err := tsuruCtx.NewRequest("POST", "/apps/"+appName+"/deploy", requestReader)
	if err != nil {
		return err
	}
	httpResponse, err := tsuruCtx.RawHTTPClient().Do(request)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNotFound {
		return fmt.Errorf("app %q not found", appName)
	}

	// ----
	debugWriter := io.Discard
	debug := tsuruCtx.Verbosity() > 0 // e.g. --verbosity 2
	if debug {
		debugWriter = tsuruCtx.Stderr
	}
	_ = debugWriter
	return nil
}
