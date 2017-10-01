package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/prattmic/restic-remote/log"
)

func tempExecutable(dir, prefix string) (*os.File, error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return f, fmt.Errorf("error creating tmpfile: %v", err)
	}

	// Must be executable.
	if err := f.Chmod(0755); err != nil {
		f.Close()
		return nil, fmt.Errorf("error setting permissions: %v", err)
	}

	return f, nil
}

func execve(bin string, args []string) {
	argv := []string{bin}
	argv = append(argv, args...)
	log.Infof("execve %v", argv)
	if err := syscall.Exec(bin, argv, os.Environ()); err != nil {
		log.Exitf("execve failed: %v", err)
	}
}
