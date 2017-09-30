package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"

	"github.com/golang/glog"
)

func deployRelease(release string) error {
	if *bucket == "" {
		return fmt.Errorf("-bucket must be set")
	}

	f := filepath.Join(release, "VERSION")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading version: %v", err)
	}
	version := string(b)

	destDir, err := url.Parse(*bucket)
	if err != nil {
		return fmt.Errorf("malformed bucket %s", *bucket)
	}
	destDir.Path = path.Join(destDir.Path, version)
	glog.Infof("Deploying version: %s to %s", version, destDir)

	return nil
}
