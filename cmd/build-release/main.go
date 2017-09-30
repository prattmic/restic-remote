package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
)

func createEmptyDir(name string) error {
	// Attempt to remove the directory. If it is not empty, this will fail.
	if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.Mkdir(name, 0755)
}

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	if err := createEmptyDir("release"); err != nil {
		glog.Exitf("Unable to create release directory: %v", err)
	}
}
