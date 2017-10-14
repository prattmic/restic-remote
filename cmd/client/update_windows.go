package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/prattmic/restic-remote/log"
)

func tempExecutable(dir, prefix string) (*os.File, error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return nil, fmt.Errorf("error creating tmpfile: %v", err)
	}

	name := f.Name()
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("error closing file: %v", err)
	}

	// Must end in .exe.
	if err := os.Rename(name, name+".exe"); err != nil {
		return nil, fmt.Errorf("error renaming file: %v", err)
	}

	f, err = os.OpenFile(name+".exe", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("error re-opening file: %v", err)
	}


	return f, nil
}

func execve(bin string, args []string, envv []string) {
	c := []string{bin}
	c = append(c, args...)
	log.Infof("Running command %v", c)

	cmd := exec.Command(c[0], c[1:]...)
	cmd.Env = envv
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Exitf("Start failed: %v", err)
	}

	log.Infof("Exiting")
	os.Exit(0)
}
