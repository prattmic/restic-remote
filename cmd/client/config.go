package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/config"
	"github.com/prattmic/restic-remote/log"
	"github.com/prattmic/restic-remote/restic"
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

// boundBoolFlag is equivalent to boundStringFlag for bool flags.
func boundBoolFlag(name string, d bool, desc string) {
	fname := strings.Replace(name, ".", "-", -1)
	pflag.Bool(fname, d, desc)
	viper.BindPFlag(name, pflag.Lookup(fname))
}

func init() {
	// viper top-level options.
	boundStringSliceFlag("backup", nil, "list of paths to backup")
	boundStringFlag("hostname", "", "hostname to use for api and snapshots")
	boundBoolFlag("update", true, "perform an update check")

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
	boundStringFlag("restic.limit-download", "", "restic download bandwidth limit (KiB/s)")
	boundStringFlag("restic.limit-upload", "", "restic upload bandwidth limit (KiB/s)")

	// viper "google" sub-tree.
	boundStringFlag("google.project-number", "", "Google Cloud project number for restic and update GCS operations")
	boundStringFlag("google.credentials", "", "Google Cloud application credentials JSON path for restic and update GCS operation")
	boundStringFlag("google.binary-bucket", "", "Bucket containing binary releases (just name, no gs://)")
}

// configFolderName is the application directory inside of the system config
// dir when our configuration is stored.
const configFolderName = "restic-remote"

// readConfig reads the global viper config.
func readConfig() {
	cd, err := config.Dir(configFolderName)
	if err != nil {
		log.Warningf("Unable to find config directory: %v", err)
	} else {
		viper.AddConfigPath(cd)
	}

	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warningf("Unable to read config: %v", err)
	}

	if viper.GetString("hostname") == "" {
		log.Exitf("hostname required")
	}
}

// newAPI creates an api.API from the viper config.
func newAPI(ctx context.Context) (*api.API, error) {
	var aconf api.Config
	if err := viper.UnmarshalKey("api", &aconf); err != nil {
		return nil, fmt.Errorf("error unmarshalling API config: %v", err)
	}
	aconf.Hostname = viper.GetString("hostname")

	return api.New(ctx, aconf)
}

// newRestic creates a restic.Restic from the viper config.
func newRestic() (*restic.Restic, error) {
	var rconf restic.Config
	if err := viper.UnmarshalKey("restic", &rconf); err != nil {
		return nil, fmt.Errorf("error unmarshalling restic config: %v", err)
	}
	rconf.Hostname = viper.GetString("hostname")
	rconf.BackendOptions = map[string]string{
		"GOOGLE_PROJECT_ID":              viper.GetString("google.project-number"),
		"GOOGLE_APPLICATION_CREDENTIALS": viper.GetString("google.credentials"),
	}

	return restic.New(rconf)
}
