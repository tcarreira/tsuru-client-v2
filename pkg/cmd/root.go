// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/config"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
	"github.com/tsuru/tsuru-client/pkg/cmd/app"
	"github.com/tsuru/tsuru-client/pkg/cmd/auth"
)

var (
	Version string = "dev"

	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tsuru",
	Short: "A command-line interface for interacting with tsuru",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		fmt.Println("This would the tsuru-plugin: " + args[0])
		fmt.Println("Not implemented yet.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tsuru/.tsuru-client.yaml)")
	rootCmd.PersistentFlags().Bool("json", false, "return the output in json format (when possible)")
	rootCmd.PersistentFlags().String("target", "", "Tsuru server endpoint")
	rootCmd.PersistentFlags().IntP("verbosity", "v", 0, "Verbosity level: 1 => print HTTP requests; 2 => print HTTP requests/responses")

	// Add subcommands
	rootCmd.AddCommand(app.NewAppCmd())
	rootCmd.AddCommand(auth.NewLoginCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
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

	SetupTsuruClientSingleton()
}

func SetupTsuruClientSingleton() {
	cfg := tsuru.NewConfiguration()
	cfg.UserAgent = "tsuru-client:" + Version

	target, err := config.GetTarget()
	cobra.CheckErr(err)
	cfg.BasePath = target

	token, err := config.GetToken()
	cobra.CheckErr(err)
	if token != "" {
		cfg.AddDefaultHeader("Authorization", "bearer "+token)
	}

	tsuructx.SetupTsuruContextSingleton(cfg, &tsuructx.TsuruContextOpts{
		InsecureSkipVerify: viper.GetBool("insecure-skip-verify"),
		Verbosity:          viper.GetInt("verbosity"),
		LocalTZ:            time.Local,
		AuthScheme:         viper.GetString("auth-scheme"),
	})
}
