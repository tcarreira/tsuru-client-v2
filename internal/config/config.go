package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ConfigPath string

func init() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	ConfigPath = filepath.Join(home, ".tsuru")
}
