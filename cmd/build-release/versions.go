package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/prattmic/restic-remote/binver"
)

type versions struct {
	release string
	restic  string
	client  string
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
	ver.restic, err = binver.Restic(filepath.Join(release, "restic"))
	if err != nil {
		return nil, fmt.Errorf("error reading restic version: %v", err)
	}

	glog.Infof("Finding client version...")
	ver.client, err = binver.Client(filepath.Join(release, "client"))
	if err != nil {
		return nil, fmt.Errorf("error reading client version: %v", err)
	}

	glog.Infof("Found versions: %+v", ver)

	return &ver, nil
}
