// Package restic provides an interface to restic backup.
package restic

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Config contains the configuration for restic.
type Config struct {
	// Binary is the path to the restic binary.
	Binary string

	// Repository is the restic repository, in the format used by restic:
	//
	// https://restic.readthedocs.io/en/latest/manual.html#initialize-a-repository
	Repository string

	// Password is the repository password.
	Password string

	// Hostname is the hostname to use for this machine when performing
	// backups.
	Hostname string

	// BackendOptions are optional options for the repository backend. They
	// are passed as environment variables to restic.
	BackendOptions map[string]string
}

// Restic provides an interface to the restic binary.
type Restic struct {
	config Config
}

// New creates a new Restic.
func New(c Config) (*Restic, error) {
	if c.Binary == "" {
		return nil, fmt.Errorf("restic binary path must be provided")
	}
	if c.Repository == "" {
		return nil, fmt.Errorf("restic repository must be provided")
	}
	if c.Password == "" {
		return nil, fmt.Errorf("restic password must be provided")
	}
	if c.Hostname == "" {
		return nil, fmt.Errorf("restic hostname must be provided")
	}

	return &Restic{
		config: c,
	}, nil
}

// run runs restic with args, returning stdout and stderr.
//
// The repository, password, hostname, and backend options are all added to the
// environment.
func (r *Restic) run(args ...string) (string, string, error) {
	c := exec.Command(r.config.Binary, args...)

	c.Env = append(c.Env, "RESTIC_REPOSITORY="+r.config.Repository)
	c.Env = append(c.Env, "RESTIC_PASSWORD="+r.config.Password)
	for k, v := range r.config.BackendOptions {
		c.Env = append(c.Env, k+"="+v)
	}

	var so, se bytes.Buffer
	c.Stdout = &so
	c.Stderr = &se

	err := c.Run()
	return so.String(), se.String(), err
}

// Version returns the complete restic version string.
func (r *Restic) Version() (string, error) {
	so, se, err := r.run("version")
	if err != nil {
		return "", fmt.Errorf("'restic version' failed with error %v. stderr: %s", err, se)
	}

	return so, nil
}

// Snapshots returns the all restic snapshots. It does no parsing.
func (r *Restic) Snapshots() (string, error) {
	so, se, err := r.run("snapshots", "--host", r.config.Hostname)
	if err != nil {
		return "", fmt.Errorf("'restic snapshot' failed with error %v. stderr: %s", err, se)
	}

	return so, nil
}
