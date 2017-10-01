package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// versionStr is the current version. It is overridden by the linker.
var versionStr = "<unknown>"

var (
	// configPath allows overriding the config file location.
	configPath = pflag.String("config", "", "Path to config file")

	// version prints the current version then exits.
	version = pflag.Bool("version", false, "Print version and exit")
)

func updateCheck(a *api.API) {
	restic, err := a.GetBinary("restic")
	if err != nil {
		log.Errorf("Error getting restic binary info: %v", err)
		return
	}

	log.Infof("restic: %+v", restic)

	remote, err := a.GetBinary("client")
	if err != nil {
		log.Errorf("Error getting client binary info: %v", err)
		return
	}

	log.Infof("client: %+v", remote)
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Trick glog into thinking we called flag.Parse.
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	if *version {
		fmt.Printf("%s\n", versionStr)
		os.Exit(0)
	}

	readConfig()

	log.Infof("restic-remote client started")

	ctx := context.Background()

	a, err := newAPI(ctx)
	if err != nil {
		log.Exitf("Failed to create API: %v", err)
	}

	if err := a.ClientStarted(); err != nil {
		log.Warningf("Error writing ClientStarted event: %v", err)
	}

	updateCheck(a)

	r, err := newRestic()
	if err != nil {
		log.Exitf("Failed to create restic: %v", err)
	}

	backup := viper.GetStringSlice("backup")
	if len(backup) < 1 {
		log.Exitf("Nothing to back up!")
	}

	log.Infof("Backing up %+v", backup)

	if err := a.BackupStarted(backup); err != nil {
		log.Warningf("Error writing BackupStarted event: %v", err)
	}

	so, se, err := r.Backup(backup)
	message := fmt.Sprintf("stdout:\n%s\nstderr:\n%s", so, se)
	log.Infof("restic backup: %s\n", message)
	if err != nil {
		if err := a.BackupFailed(message); err != nil {
			log.Warningf("Error writing BackupFailed event: %v", err)
		}
		log.Exitf("Failed to backup: %v", err)
	}

	if err := a.BackupSucceeded(message); err != nil {
		log.Warningf("Error writing BackupSucceeded event: %v", err)
	}
}
