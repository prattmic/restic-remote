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

	// BackupStarted indicates that a backup has begun.
	BackupStarted Type = "backup_started"

	// BackupComplete indicates that a backup is complete.
	BackupComplete Type = "backup_complete"
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
