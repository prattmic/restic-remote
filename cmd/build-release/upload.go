package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

func checkNotExist(dir string) error {
	c := []string{"gsutil", "ls", dir}
	glog.Infof("Running command: %v", c)

	cmd := exec.Command(c[0], c[1:]...)
	b, err := cmd.CombinedOutput()
	if err == nil {
		return fmt.Errorf("%s exists; output: %s", dir, string(b))
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return fmt.Errorf("err %v is not ExitError; output: %s", err, string(b))
	}
	ws, ok := ee.Sys().(syscall.WaitStatus)
	if !ok {
		return fmt.Errorf("ee %v does not contain WaitStatus; output: %s", ee, string(b))
	}
	if ws.ExitStatus() != 1 {
		return fmt.Errorf("unexpected exit code %+v; output: %s", ws, string(b))
	}
	// Exit code 1 means directory doesn't exist.
	return nil
}

func gsCopy(dst string, src ...string) error {
	c := []string{"gsutil", "cp"}
	c = append(c, src...)
	c = append(c, dst)
	glog.Infof("Running command: %v", c)

	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %v failed: %v", c, err)
	}
	return nil
}

func uploadRelease(release string, ver *versions) error {
	bucket := viper.GetString("bucket")
	if bucket == "" {
		return fmt.Errorf("bucket must be set")
	}

	destDir, err := url.Parse(bucket)
	if err != nil {
		return fmt.Errorf("malformed bucket %s", bucket)
	}
	destDir.Path = path.Join(destDir.Path, ver.release)
	glog.Infof("Deploying version: %s to %s", ver.release, destDir)

	if err := checkNotExist(destDir.String()); err != nil {
		return fmt.Errorf("%s not empty: %v", destDir, err)
	}

	srcs := []string{
		filepath.Join(release, "restic"),
		filepath.Join(release, "restic.exe"),
		filepath.Join(release, "client"),
		filepath.Join(release, "client.exe"),
	}

	if err := gsCopy(destDir.String(), srcs...); err != nil {
		return fmt.Errorf("error copying: %v", err)
	}

	return nil
}
