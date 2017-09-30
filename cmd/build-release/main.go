package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	build = flag.Bool("build", true, "build new release")
)

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		glog.Exitf("Unable to working directory: %v", err)
	}

	release := filepath.Join(root, "release")

	if *build {
		if err := buildRelease(root, release); err != nil {
			glog.Errorf("Unable to build release: %v", err)
		}
	}
}
