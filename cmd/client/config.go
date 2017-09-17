package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// boundStringFlag defines a new flag that is bound to a viper key.
//
// name is the viper key name. The flag name replaces . with -.
func boundStringFlag(name, d, desc string) {
	fname := strings.Replace(name, ".", "-", -1)
	pflag.String(fname, d, desc)
	viper.BindPFlag(name, pflag.Lookup(fname))
}

// boundStringSliceFlag is equivalent to boundStringFlag for StringSlice flags.
func boundStringSliceFlag(name string, d []string, desc string) {
	fname := strings.Replace(name, ".", "-", -1)
	pflag.StringSlice(fname, d, desc)
	viper.BindPFlag(name, pflag.Lookup(fname))
}

func init() {
	// viper top-level options.
	boundStringSliceFlag("backup", nil, "list of paths to backup")
	boundStringFlag("hostname", "", "hostname to use for api and snapshots")

	// viper "api" sub-tree.
	boundStringFlag("api.root", "", "API root URL")
	boundStringFlag("api.client-id", "", "API client ID")
	boundStringFlag("api.client-secret", "", "API client secret")
	boundStringFlag("api.audience", "", "API audience name")
	boundStringFlag("api.token-url", "", "API token URL")

	// viper "restic" sub-tree.
	boundStringFlag("restic.binary", "", "restic binary path")
	boundStringFlag("restic.repository", "", "restic repository path")
	boundStringFlag("restic.password", "", "restic repository password")
	boundStringFlag("restic.google-project-id", "", "Google Cloud project ID for restic repositories using the GCS backend")
	boundStringFlag("restic.google-credentials", "", "Google Cloud application credentials JSON path for restic repositories using the GCS backend")
}

// configFolderName is the application directory inside of the system config
// dir when our configuration is stored.
const configFolderName = "restic-remote"

// configDir returns the default configuration directory.
func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData, ok := os.LookupEnv("APPDATA")
		if !ok {
			return "", fmt.Errorf("APPDATA not set")
		}

		return filepath.Join(appData, configFolderName), nil
	}

	xdg, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if ok {
		return filepath.Join(xdg, configFolderName), nil
	}

	home, ok := os.LookupEnv("HOME")
	if ok {
		return filepath.Join(home, ".config", configFolderName), nil
	}

	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot determine current user: %v", err)
	}

	if u.HomeDir != "" {
		return filepath.Join(u.HomeDir, ".config", configFolderName), nil
	}

	return "", fmt.Errorf("unable to find config directory")
}

// readConfig reads the global viper config.
func readConfig() {
	cd, err := configDir()
	if err != nil {
		log.Printf("Unable to find config directory: %v", err)
	} else {
		viper.AddConfigPath(cd)
	}

	if *config != "" {
		viper.SetConfigFile(*config)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Unable to read config: %v", err)
	}
}
