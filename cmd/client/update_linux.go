package main

import (
	"os"
	"syscall"

	"github.com/prattmic/restic-remote/log"
)

func execve(bin string, args []string) {
	argv := []string{bin}
	argv = append(argv, args...)
	log.Infof("execve %v", argv)
	if err := syscall.Exec(bin, argv, os.Environ()); err != nil {
		log.Exitf("execve failed: %v", err)
	}
}
