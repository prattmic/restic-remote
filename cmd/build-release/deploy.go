package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/golang/glog"
)

func gsCopy(dst, src string) error {
	c := []string{"gsutil", "cp", "-n", src, dst}
	glog.Infof("Running command: %v", c)

	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %v failed: %v", c, err)
	}
	return nil
}

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

	fn := func(bin string) error {
		src := filepath.Join(release, bin)
		dst := *destDir
		dst.Path = path.Join(dst.Path, bin)
		glog.Infof("Copying %s to %s", src, dst)
		return gsCopy(dst.String(), src)
	}

	if err := fn("restic"); err != nil {
		return fmt.Errorf("error copying restic: %v", err)
	}
	if err := fn("restic.exe"); err != nil {
		return fmt.Errorf("error copying restic: %v", err)
	}
	if err := fn("client"); err != nil {
		return fmt.Errorf("error copying restic: %v", err)
	}
	if err := fn("client.exe"); err != nil {
		return fmt.Errorf("error copying restic: %v", err)
	}

	return nil
}
