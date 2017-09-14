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
	"time"

	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/event"
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

func init() {
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
	boundStringFlag("restic.hostname", "", "hostname to use for restic snapshots")
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

	var aconf api.Config
	if err := viper.UnmarshalKey("api", &aconf); err != nil {
		log.Fatalf("Error unmarshalling API config: %v", err)
	}

	a, err := api.New(context.Background(), aconf)
	if err != nil {
		log.Fatalf("Failed to create API: %v", err)
	}

	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error retrieving hostname: %v", err)
	}

	if err := a.WriteEvent(&event.Event{
		Type:      event.ClientStarted,
		Timestamp: time.Now(),
		Hostname:  host,
		Message:   "Hello world",
	}); err != nil {
		log.Fatalf("Error writing event: %v", err)
	}

	var rconf restic.Config
	if err := viper.UnmarshalKey("restic", &rconf); err != nil {
		log.Fatalf("Error unmarshalling restic config: %v", err)
	}

	rconf.BackendOptions = map[string]string{
		"GOOGLE_PROJECT_ID":              viper.GetString("restic.google-project-id"),
		"GOOGLE_APPLICATION_CREDENTIALS": viper.GetString("restic.google-credentials"),
	}

	r, err := restic.New(rconf)
	if err != nil {
		log.Fatalf("Failed to create restic: %v", err)
	}

	v, err := r.Snapshots()
	if err != nil {
		log.Fatalf("Failed to get snapshots: %v", err)
	}

	log.Printf("restic snapshots: %s", v)
}
