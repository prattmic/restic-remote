package main

import (
	"os"

	"github.com/prattmic/restic-remote/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// wrapRestic exec's restic with the client config.
func wrapRestic() {
	bin := viper.GetString("restic.binary")
	if bin == "" {
		log.Exitf("restic binary path missing")
	}

	envv := os.Environ()
	envv = append(envv, "RESTIC_REPOSITORY="+viper.GetString("restic.repository"))
	envv = append(envv, "RESTIC_PASSWORD="+viper.GetString("restic.password"))
	envv = append(envv, "GOOGLE_PROJECT_ID="+viper.GetString("google.project-number"))
	envv = append(envv, "GOOGLE_APPLICATION_CREDENTIALS="+viper.GetString("google.credentials"))

	execve(bin, pflag.Args(), envv)

	log.Exitf("execve returned")
}
