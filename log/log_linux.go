package log

import (
	"log/syslog"
)

var writer *syslog.Writer

func init() {
	var err error
	writer, err = syslog.New(syslog.LOG_INFO, logTag)
	if err != nil {
		Errorf("Failed to connect to syslog: %v", err)
	}
}

func systemInfo(message string) {
	if writer == nil {
		return
	}
	writer.Info(message)
}

func systemWarning(message string) {
	if writer == nil {
		return
	}
	writer.Warning(message)
}

func systemError(message string) {
	if writer == nil {
		return
	}
	writer.Err(message)
}
