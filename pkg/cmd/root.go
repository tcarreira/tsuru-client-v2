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
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "tsuru",
		Short: "A command-line interface for interacting with tsuru",
	}

	// Flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tsuru/.tsuru-client.yaml)")
	rootCmd.PersistentFlags().Bool("json", false, "return the output in json format (when possible)")
	rootCmd.PersistentFlags().String("target", "", "Tsuru server endpoint")
	rootCmd.PersistentFlags().IntP("verbosity", "v", 0, "Verbosity level: 1 => print HTTP requests; 2 => print HTTP requests/responses")

	// setupConfig (parse configFile and bind environment variables)
	setupConfig(rootCmd)

	// Add subcommands
	rootCmd.AddCommand(app.NewAppCmd())
	rootCmd.AddCommand(auth.NewLoginCmd())
	rootCmd.AddCommand(auth.NewLogoutCmd())

	SetupTsuruClientSingleton()
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { // only after SetupTsuruClientSingleton()
		return runTsuruPluginOrHelp(cmd, args, tsuructx.GetTsuruContextSingleton())
	}
	return rootCmd
}

func runTsuruPluginOrHelp(cmd *cobra.Command, args []string, tsuruCtx *tsuructx.TsuruContext) error {
	if len(args) == 0 {
		cmd.SetOutput(tsuruCtx.Stdout)
		cmd.Help()
		return nil
	}
	fmt.Fprintln(tsuruCtx.Stdout, "This would the tsuru-plugin: "+strings.Join(args[0:], " "))
	fmt.Fprintln(tsuruCtx.Stdout, "Not implemented yet.")
	return nil
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

func SetupTsuruClientSingleton() {
	osFs := afero.NewOsFs()
	var err error

	// Get target
	target := viper.GetString("target")
	if target == "" {
		target, err = config.GetCurrentTargetFromFs(osFs)
		cobra.CheckErr(err)
	}
	target, err = config.GetTargetURL(osFs, target)
	cobra.CheckErr(err)

	// Get token
	token := viper.GetString("token")
	if token == "" {
		token, err = config.GetTokenFromFs(osFs)
		cobra.CheckErr(err)
	}

	tsuructx.SetupTsuruContextSingleton(productionOpts(osFs, token, target))
}

func productionOpts(fs afero.Fs, token, target string) *tsuructx.TsuruContextOpts {
	return &tsuructx.TsuruContextOpts{
		Verbosity:          viper.GetInt("verbosity"),
		InsecureSkipVerify: viper.GetBool("insecure-skip-verify"),
		LocalTZ:            time.Local,
		AuthScheme:         viper.GetString("auth-scheme"),
		Executor:           &exec.OsExec{},
		Fs:                 fs,

		UserAgent: "tsuru-client:" + Version,
		TargetURL: target,
		Token:     token,

		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}
