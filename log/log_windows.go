package log

import (
	"golang.org/x/sys/windows/svc/eventlog"
)

var log *eventlog.Log

func init() {
	var err error
	log, err = eventlog.Open(logTag)
	if err != nil {
		Errorf("Failed to open event log: %v", err)
	}
}

func systemInfo(message string) {
	if log == nil {
		return
	}
	log.Info(1, message)
}

func systemWarning(message string) {
	if log == nil {
		return
	}
	log.Warning(1, message)
}

func systemError(message string) {
	if log == nil {
		return
	}
	log.Error(1, message)
}
