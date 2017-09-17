// Package log logs to stderr and the system log.
package log

import (
	"flag"

	"github.com/golang/glog"
)

func init() {
	// Just log glog to stderr.
	flag.Set("logtostderr", "true")
}

// Infof logs at the info level.
func Infof(format string, args ...interface{}) {
	glog.Infof(format, args...)
}

// Warningf logs at the warning level.
func Warningf(format string, args ...interface{}) {
	glog.Warningf(format, args...)
}

// Errorf logs at the error level.
func Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

// Exitf logs at the error level, then exits with a non-zero exit code.
func Exitf(format string, args ...interface{}) {
	glog.Exitf(format, args...)
}

// Fatalf logs at the error level, then dumps a stack trace and exits with a
// non-zero exit code.
func Fatalf(format string, args ...interface{}) {
	glog.Fatalf(format, args...)
}
