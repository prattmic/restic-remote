// Package log logs to stderr and the system log.
package log

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
)

// logTag is used to identify logs from this package in the system log.
const logTag = "restic-remote"

func init() {
	// Just log glog to stderr.
	flag.Set("logtostderr", "true")
}

// Infof logs at the info level.
func Infof(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	systemInfo(m)
	glog.InfoDepth(1, m)
}

// Warningf logs at the warning level.
func Warningf(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	systemWarning(m)
	glog.WarningDepth(1, m)
}

// Errorf logs at the error level.
func Errorf(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	systemError(m)
	glog.ErrorDepth(1, m)
}

// Exitf logs at the error level, then exits with a non-zero exit code.
func Exitf(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	systemError(m)
	glog.ExitDepth(1, m)
}

// Fatalf logs at the error level, then dumps a stack trace and exits with a
// non-zero exit code.
func Fatalf(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	systemError(m)
	glog.FatalDepth(1, m)
}
