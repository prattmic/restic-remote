package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
)

type versions struct {
	release string
	restic  string
	client  string
}

func clientVersion(bin string) (string, error) {
	cmd := exec.Command(bin, "--version")
	b, err := cmd.Output()
	return string(b), err
}

func resticVersion(bin string) (string, error) {
	cmd := exec.Command(bin, "version")
	b, err := cmd.Output()
	return string(b), err
}

func findVersions(release string) (*versions, error) {
	var ver versions

	glog.Infof("Finding release version...")
	f := filepath.Join(release, "VERSION")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("error reading release version: %v", err)
	}
	ver.release = string(b)

	glog.Infof("Finding restic version...")
	ver.restic, err = resticVersion(filepath.Join(release, "restic"))
	if err != nil {
		return nil, fmt.Errorf("error reading restic version: %v", err)
	}

	glog.Infof("Finding client version...")
	ver.client, err = clientVersion(filepath.Join(release, "client"))
	if err != nil {
		return nil, fmt.Errorf("error reading client version: %v", err)
	}

	glog.Infof("Found versions: %+v", ver)

	return &ver, nil
}
