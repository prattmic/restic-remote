package main

import (
	"os"
	"os/exec"

	"github.com/prattmic/restic-remote/log"
)

func execve(bin string, args []string) {
	c := []string{bin}
	c = append(c, args...)
	log.Infof("Running command %v", c)

	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Exitf("Start failed: %v", err)
	}

	log.Infof("Exiting")
	os.Exit(0)
}
