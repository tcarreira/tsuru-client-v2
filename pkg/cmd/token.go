package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func getToken(fsys afero.Fs) (string, error) {
	var token []byte
	if token := viper.GetString("token"); token != "" {
		return token, nil
	}

	tokenPaths := []string{filepath.Join(configPath, "token")}
	if targetLabel, err := getTargetLabel(fsys); err == nil {
		tokenPaths = append([]string{filepath.Join(configPath, "token.d", targetLabel)}, tokenPaths...)
	}

	var err error
	for _, tokenPath := range tokenPaths {
		var tkFile afero.File
		if tkFile, err = fsys.Open(tokenPath); err == nil {
			defer tkFile.Close()
			token, err = io.ReadAll(tkFile)
			if err != nil {
				return "", err
			}
			tokenStr := strings.TrimSpace(string(token))
			return tokenStr, nil
		}
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return "", err
}

// GetToken returns the token for the current target,
// as defined in the TSURU_TOKEN environment variable or in the token file.
func GetToken() (string, error) {
	return getToken(afero.NewOsFs())
}
