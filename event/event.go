// Package event defines backup events.
package event

import (
	"time"
)

// Type describes the type of event.
type Type string

const (
	// ClientStarted indicates that a client has just come online.
	ClientStarted Type = "client_started"

	// ClientVersion contains the version of the client.
	ClientVersion Type = "client_version"

	// ResticVersion contains the version of restic.
	ResticVersion Type = "restic_version"

	// UpdateComplete indicates that the release has been updated and the
	// client is restarting.
	UpdateComplete Type = "update_complete"

	// BackupStarted indicates that a backup has begun.
	BackupStarted Type = "backup_started"

	// BackupSucceeded indicates that a backup completed successfully.
	BackupSucceeded Type = "backup_succeeded"

	// BackupFailed indicates that a backup completed unsuccessfully.
	BackupFailed Type = "backup_failed"
)

// Event describes a single event.
type Event struct {
	// Type is one of the above constants.
	Type Type

	// Timestamp is the time of the event.
	Timestamp time.Time

	// Hostname is the machine on which the even occurred.
	Hostname string

	// Message is an optional free-form message.
	Message string
}
