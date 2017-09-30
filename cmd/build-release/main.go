package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

func createEmptyDir(name string) error {
	// Attempt to remove the directory. If it is not empty, this will fail.
	if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.Mkdir(name, 0755)
}

func buildRestic(root, release string) error {
	bin := filepath.Join(release, "restic")
	glog.Infof("Building %s", bin)

	cmd := exec.Command("go", "run", "build.go", "--verbose", "--output", bin)
	cmd.Dir = filepath.Join(root, "tools", "restic")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building restic: %v", err)
	}

	bin = filepath.Join(release, "restic.exe")
	glog.Infof("Building %s", bin)

	cmd = exec.Command("go", "run", "build.go", "--verbose", "--goos", "windows", "--output", bin)
	cmd.Dir = filepath.Join(root, "tools", "restic")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building restic.exe: %v", err)
	}

	return nil
}

func buildClient(root, release string) error {
	glog.Infof("Determing client version...")

	cmd := exec.Command("git", "describe", "--long", "--tags", "--dirty", "--always")
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error finding version: %v", err)
	}

	version := strings.Trim(string(b), "\r\n")
	glog.Infof("Client version: %s", version)

	ldflag := fmt.Sprintf(`-X "main.versionStr=%s"`, version)

	bin := filepath.Join(release, "client")
	glog.Infof("Building %s", bin)

	cmd = exec.Command("go", "build", "-o", bin, "-ldflags", ldflag, "github.com/prattmic/restic-remote/cmd/client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building client: %v", err)
	}

	bin = filepath.Join(release, "client.exe")
	glog.Infof("Building %s", bin)

	cmd = exec.Command("go", "build", "-o", bin, "-ldflags", ldflag, "github.com/prattmic/restic-remote/cmd/client")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=windows")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building client.exe: %v", err)
	}

	return nil
}

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		glog.Exitf("Unable to working directory: %v", err)
	}

	release := filepath.Join(root, "release")

	if err := createEmptyDir(release); err != nil {
		glog.Exitf("Unable to create release directory: %v", err)
	}

	if err := buildRestic(root, release); err != nil {
		glog.Exitf("Unable to build restic: %v", err)
	}

	if err := buildClient(root, release); err != nil {
		glog.Exitf("Unable to build client: %v", err)
	}
}
