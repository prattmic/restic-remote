package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/restic"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// config allows overriding the config file location.
	config = pflag.String("config", "", "Path to config file")
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

const configFolderName = "restic-remote"

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

func main() {
	pflag.Parse()

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

	hostname := viper.GetString("hostname")
	if hostname == "" {
		log.Fatalf("hostname required")
	}

	var aconf api.Config
	if err := viper.UnmarshalKey("api", &aconf); err != nil {
		log.Fatalf("Error unmarshalling API config: %v", err)
	}
	aconf.Hostname = hostname

	a, err := api.New(context.Background(), aconf)
	if err != nil {
		log.Fatalf("Failed to create API: %v", err)
	}

	if err := a.ClientStarted(); err != nil {
		log.Printf("Error writing ClientStarted event: %v", err)
	}

	var rconf restic.Config
	if err := viper.UnmarshalKey("restic", &rconf); err != nil {
		log.Fatalf("Error unmarshalling restic config: %v", err)
	}
	rconf.Hostname = hostname
	rconf.BackendOptions = map[string]string{
		"GOOGLE_PROJECT_ID":              viper.GetString("restic.google-project-id"),
		"GOOGLE_APPLICATION_CREDENTIALS": viper.GetString("restic.google-credentials"),
	}

	r, err := restic.New(rconf)
	if err != nil {
		log.Fatalf("Failed to create restic: %v", err)
	}

	backup := viper.GetStringSlice("backup")
	if len(backup) < 1 {
		log.Fatalf("Nothing to back up!")
	}

	if err := a.BackupStarted(backup); err != nil {
		log.Printf("Error writing BackupStarted event: %v", err)
	}

	so, se, err := r.Backup(backup)
	message := fmt.Sprintf("stdout:\n%s\nstderr:\n%s", so, se)
	log.Printf("restic backup: %s\n", message)
	if err != nil {
		if err := a.BackupFailed(message); err != nil {
			log.Printf("Error writing BackupFailed event: %v", err)
		}
		log.Fatalf("Failed to backup: %v", err)
	}

	if err := a.BackupSucceeded(message); err != nil {
		log.Printf("Error writing BackupSucceeded event: %v", err)
	}
}
