package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	build  = flag.Bool("build", true, "build new release")
	deploy = flag.Bool("deploy", false, "deploy new release")

	bucket = flag.String("bucket", "", "bucket to deploy to (gs://foo/)")
)

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		glog.Exitf("Unable to working directory: %v", err)
	}

	release := filepath.Join(root, "release")

	var ver *versions
	if *build {
		var err error
		ver, err = buildRelease(root, release)
		if err != nil {
			glog.Exitf("Unable to build release: %v", err)
		}
	}

	if ver == nil {
		var err error
		ver, err = findVersions(release)
		if err != nil {
			glog.Exitf("Unable to determine versions: %v", err)
		}
	}

	if *deploy {
		if err := deployRelease(release, ver); err != nil {
			glog.Exitf("Unable to deploy release: %v", err)
		}
	}
}
