package main

import (
	"os"
	"os/exec"

	"github.com/prattmic/restic-remote/log"
)

func execve(bin string, args []string) {
	log.Infof("Running command %v", argv)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Exitf("Start failed: %v", err)
	}

	log.Infof("Exiting")
	os.Exit(0)
}
