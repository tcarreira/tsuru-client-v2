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

var commands = []func(*tsuructx.TsuruContext) *cobra.Command{
	app.NewAppCmd,
	auth.NewLoginCmd,
	auth.NewLogoutCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := newRootCmd(viper.GetViper(), nil)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func newRootCmd(vip *viper.Viper, tsuruCtx *tsuructx.TsuruContext) *cobra.Command {
	vip = preSetupViper(vip)
	if tsuruCtx == nil {
		tsuruCtx = NewProductionTsuruContext(vip, afero.NewOsFs())
	}
	rootCmd := newBareRootCmd(tsuruCtx)
	setupPFlagsAndCommands(rootCmd, tsuruCtx)
	return rootCmd
}

func runRootCmd(tsuruCtx *tsuructx.TsuruContext, cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	if len(args) == 0 {
		cmd.SetOut(tsuruCtx.Stdout)
		cmd.SetErr(tsuruCtx.Stderr)
		return cmd.Help()
	}

	pluginName := args[0]
	if tsuruCtx.Viper.GetString("plugin-name") == pluginName {
		return fmt.Errorf("failing trying to run recursive plugin")
	}

	pluginPath := findExecutablePlugin(tsuruCtx, pluginName)
	if pluginPath == "" {
		return fmt.Errorf("command not found")
	}

	envs := os.Environ()
	tsuruEnvs := []string{
		"TSURU_TARGET=" + tsuruCtx.TargetURL(),
		"TSURU_TOKEN=" + tsuruCtx.Token(),
		"TSURU_VERBOSITY=" + fmt.Sprintf("%d", tsuruCtx.Verbosity()),
		"TSURU_PLUGIN_NAME=" + pluginName,
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

func rootPersistentPreRun(tsuruCtx *tsuructx.TsuruContext) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if l := cmd.Flags().Lookup("target"); l != nil && l.Value.String() != "" {
			fmt.Println("debug: setting target", cmd.Flag("target").Value.String())
			tsuruCtx.SetTargetURL(l.Value.String())
		}
		if v, err := cmd.Flags().GetInt("verbosity"); err != nil {
			fmt.Println("debug: setting verbosity")
			tsuruCtx.SetVerbosity(v)
		}
	}
}

func newBareRootCmd(tsuruCtx *tsuructx.TsuruContext) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tsuru",
		Short: "A command-line interface for interacting with tsuru",

		PersistentPreRun: rootPersistentPreRun(tsuruCtx),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRootCmd(tsuruCtx, cmd, args)
		},
		Args: cobra.MinimumNArgs(0),

		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		DisableFlagParsing: true,
	}
	rootCmd.SetIn(tsuruCtx.Stdin)
	rootCmd.SetOut(tsuruCtx.Stdout)
	rootCmd.SetErr(tsuruCtx.Stderr)

	return rootCmd
}

// preSetupViper is supposed to be called before NewProductionTsuruContext()
func preSetupViper(vip *viper.Viper) *viper.Viper {
	vip.SetEnvPrefix("tsuru")
	vip.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	vip.AutomaticEnv() // read in environment variables that match
	return vip
}

// setupPFlagsAndCommands reads in config file and ENV variables if set.
func setupPFlagsAndCommands(rootCmd *cobra.Command, tsuruCtx *tsuructx.TsuruContext) {
	// Persistent Flags.
	// !!! Double bind them inside PersistentPreRun() !!!
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tsuru/.tsuru-client.yaml)")
	rootCmd.PersistentFlags().Bool("json", false, "return the output in json format (when possible)") // TODO: add to PersistentPreRun()
	rootCmd.PersistentFlags().String("target", "", "Tsuru server endpoint")
	rootCmd.PersistentFlags().IntP("verbosity", "v", 0, "Verbosity level: 1 => print HTTP requests; 2 => print HTTP requests/responses")

	if cfgFile != "" {
		// Use config file from the flag.
		tsuruCtx.Viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".tsuru-client" (without extension).
		tsuruCtx.Viper.AddConfigPath(config.ConfigPath)
		tsuruCtx.Viper.SetConfigType("yaml")
		tsuruCtx.Viper.SetConfigName(".tsuru-client")
	}

	tsuruCtx.Viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	tsuruCtx.Viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))

	// If a config file is found, read it in.
	if err := tsuruCtx.Viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", tsuruCtx.Viper.ConfigFileUsed()) // TODO: handle this better
	}

	// Add subcommands
	for _, cmd := range commands {
		rootCmd.AddCommand(cmd(tsuruCtx))
	}
}

func NewProductionTsuruContext(vip *viper.Viper, fs afero.Fs) *tsuructx.TsuruContext {
	var err error

	// Get target
	target := vip.GetString("target")
	if target == "" {
		target, err = config.GetCurrentTargetFromFs(fs)
		cobra.CheckErr(err)
	}
	target, err = config.GetTargetURL(fs, target)
	cobra.CheckErr(err)
	vip.Set("target", target)

	// Get token
	token := vip.GetString("token")
	if token == "" {
		token, err = config.GetTokenFromFs(fs)
		cobra.CheckErr(err)
	}
	vip.Set("token", token)

	return tsuructx.TsuruContextWithConfig(productionOpts(fs, vip))
}

func productionOpts(fs afero.Fs, vip *viper.Viper) *tsuructx.TsuruContextOpts {
	return &tsuructx.TsuruContextOpts{
		InsecureSkipVerify: vip.GetBool("insecure-skip-verify"),
		LocalTZ:            time.Local,
		AuthScheme:         vip.GetString("auth-scheme"),
		Executor:           &exec.OsExec{},
		Fs:                 fs,
		Viper:              vip,

		UserAgent: "tsuru-client:" + Version,

		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}
