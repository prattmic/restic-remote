package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
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

var (
	build   = pflag.Bool("build", true, "build new release")
	upload  = pflag.Bool("upload", false, "upload new release")
	rollout = pflag.Bool("rollout", false, "rollout new release")

	configPath = pflag.String("config", "", "Path to config file")
)

func init() {
	boundStringFlag("bucket", "", "bucket to upload to (gs://foo/)")

	boundStringFlag("api.root", "", "API root URL")
	boundStringFlag("api.client-id", "", "API client ID")
	boundStringFlag("api.client-secret", "", "API client secret")
	boundStringFlag("api.audience", "", "API audience name")
	boundStringFlag("api.token-url", "", "API token URL")
}

func main() {
	flag.Set("alsologtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Trick glog into thinking we called flag.Parse.
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		glog.Warningf("Unable to read config: %v", err)
	}

	root, err := os.Getwd()
	if err != nil {
		glog.Exitf("Unable to working directory: %v", err)
	}

	release := filepath.Join(root, "release")

	var ver *versions
	if *build {
		var err error
		ver, err = buildRelease(root, release)
		if err != nil {
			glog.Exitf("Unable to build release: %v", err)
		}
	}

	if ver == nil {
		var err error
		ver, err = findVersions(release)
		if err != nil {
			glog.Exitf("Unable to determine versions: %v", err)
		}
	}

	if *upload {
		if err := uploadRelease(release, ver); err != nil {
			glog.Exitf("Unable to upload release: %v", err)
		}
	}

	if *rollout {
		if err := rolloutRelease(release, ver); err != nil {
			glog.Exitf("Unable to rollout release: %v", err)
		}
	}
}
