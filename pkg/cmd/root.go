// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tsuru/tsuru-client/internal/config"
	"github.com/tsuru/tsuru-client/internal/exec"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
	"github.com/tsuru/tsuru-client/pkg/cmd/app"
	"github.com/tsuru/tsuru-client/pkg/cmd/auth"
)

var (
	Version string = "dev"
	cfgFile string
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tsuru",
		Short: "A command-line interface for interacting with tsuru",
		RunE:  func(cmd *cobra.Command, args []string) error { return fmt.Errorf("placeholder") },
	}

	// Setup cli
	setupConfig(rootCmd)
	SetupTsuruContextSingleton(viper.GetViper())
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { // only after SetupTsuruClientSingleton()
		return runRootCmd(cmd, args, tsuructx.GetTsuruContextSingleton())
	}

	// Add subcommands
	rootCmd.AddCommand(app.NewAppCmd())
	rootCmd.AddCommand(auth.NewLoginCmd())
	rootCmd.AddCommand(auth.NewLogoutCmd())

	return rootCmd
}

func runRootCmd(cmd *cobra.Command, args []string, tsuruCtx *tsuructx.TsuruContext) error {
	if len(args) == 0 {
		cmd.SetOut(tsuruCtx.Stdout)
		cmd.SetErr(tsuruCtx.Stderr)
		return cmd.Help()
	}

	pluginName := args[0]
	if viper.GetString("plugin-name") == pluginName {
		return fmt.Errorf("failing trying to run recursive plugin")
	}

	pluginPath := findExecutablePlugin(tsuruCtx, pluginName)
	if pluginPath == "" {
		return fmt.Errorf("command not found")
	}

	envs := os.Environ()
	tsuruEnvs := []string{
		"TSURU_TARGET=" + tsuruCtx.TargetURL,
		"TSURU_TOKEN=" + tsuruCtx.Token,
		"TSURU_PLUGIN_NAME=" + pluginName,
		"TSURU_VERBOSITY=" + fmt.Sprintf("%d", tsuruCtx.Verbosity),
	}
	envs = append(envs, tsuruEnvs...)

	opts := exec.ExecuteOptions{
		Cmd:    pluginPath,
		Args:   args[1:],
		Stdout: tsuruCtx.Stdout,
		Stderr: tsuruCtx.Stderr,
		Stdin:  tsuruCtx.Stdin,
		Envs:   envs,
	}
	return tsuruCtx.Executor.Command(opts)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// setupConfig reads in config file and ENV variables if set.
func setupConfig(rootCmd *cobra.Command) {
	// Persistent Flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tsuru/.tsuru-client.yaml)")
	rootCmd.PersistentFlags().Bool("json", false, "return the output in json format (when possible)")
	rootCmd.PersistentFlags().String("target", "", "Tsuru server endpoint")
	rootCmd.PersistentFlags().IntP("verbosity", "v", 0, "Verbosity level: 1 => print HTTP requests; 2 => print HTTP requests/responses")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".tsuru-client" (without extension).
		viper.AddConfigPath(config.ConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tsuru-client")
	}

	viper.SetEnvPrefix("tsuru")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed()) // TODO: handle this better
	}
}

func SetupTsuruContextSingleton(vip *viper.Viper) {
	osFs := afero.NewOsFs()
	var err error

	// Get target
	target := vip.GetString("target")
	if target == "" {
		target, err = config.GetCurrentTargetFromFs(osFs)
		cobra.CheckErr(err)
	}
	target, err = config.GetTargetURL(osFs, target)
	cobra.CheckErr(err)

	// Get token
	token := vip.GetString("token")
	if token == "" {
		token, err = config.GetTokenFromFs(osFs)
		cobra.CheckErr(err)
	}

	tsuructx.SetupTsuruContextSingleton(productionOpts(osFs, token, target, vip))
}

func productionOpts(fs afero.Fs, token, target string, vip *viper.Viper) *tsuructx.TsuruContextOpts {
	return &tsuructx.TsuruContextOpts{
		Verbosity:          vip.GetInt("verbosity"),
		InsecureSkipVerify: vip.GetBool("insecure-skip-verify"),
		LocalTZ:            time.Local,
		AuthScheme:         vip.GetString("auth-scheme"),
		Executor:           &exec.OsExec{},
		Fs:                 fs,
		Viper:              vip,

		UserAgent: "tsuru-client:" + Version,
		TargetURL: target,
		Token:     token,

		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}
