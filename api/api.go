// Package api provides an interface to the backup API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/prattmic/restic-remote/auth0"
	"github.com/prattmic/restic-remote/event"
)

// API endpoints.
const (
	eventEndpoint  = "/api/v1/event"
	binaryEndpoint = "/api/v1/binary"
)

type Config struct {
	// ClientConfig contains the authentication configuration.
	auth0.ClientConfig `mapstructure:",squash"`

	// Root is the API root URL.
	Root string `mapstructure:"root"`

	// Hostname is the hostname to use for the event helper methods.
	Hostname string
}

// API describes a configured API target.
type API struct {
	// root is the API root URL.
	//
	// root is immutable.
	root url.URL

	// hostname is the hostname to use for the event helper methods.
	hostname string

	// client connects to the API with authentication.
	client *http.Client
}

// New creates an API.
//
// ctx is used for authentication token refreshes.
func New(ctx context.Context, conf Config) (*API, error) {
	if conf.Root == "" {
		return nil, fmt.Errorf("API root must be provided")
	}

	u, err := url.Parse(conf.Root)
	if err != nil {
		return nil, fmt.Errorf("malformed root URL %q: %v", conf.Root, err)
	}

	client, err := auth0.NewClient(ctx, conf.ClientConfig)
	if err != nil {
		return nil, err
	}

	return &API{
		root:     *u,
		hostname: conf.Hostname,
		client:   client,
	}, nil
}

// url returns the full URL for the given endpoint.
//
// N.B. u.User must not be modified, as it is not deep copied.
func (a *API) url(e string) url.URL {
	u := a.root
	u.Path = path.Join(u.Path, eventEndpoint)
	return u
}

// postJSON posts JSON object j to u.
func (a *API) postJSON(u url.URL, j interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(j); err != nil {
		return fmt.Errorf("error encoding %+v: %v", j, err)
	}

	r, err := a.client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode >= 200 && r.StatusCode < 300 {
		// Success.
		return nil
	}

	// Read any error message sent.
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading body of failure response %+v: %v", r, err)
	}

	return fmt.Errorf("error response when writing JSON: %+v\n%s", r, string(b))
}

// WriteEvent sends an event to the server.
func (a *API) WriteEvent(e *event.Event) error {
	if err := a.postJSON(a.url(eventEndpoint), e); err != nil {
		return fmt.Errorf("error writing event %+v: %v", e, err)
	}
	return nil
}

// ClientStarted writes a ClientStarted event.
func (a *API) ClientStarted() error {
	return a.WriteEvent(&event.Event{
		Type:      event.ClientStarted,
		Timestamp: time.Now(),
		Hostname:  a.hostname,
	})
}

// BackupStarted writes a BackupStarted event for dirs.
func (a *API) BackupStarted(dirs []string) error {
	return a.WriteEvent(&event.Event{
		Type:      event.BackupStarted,
		Timestamp: time.Now(),
		Hostname:  a.hostname,
		Message:   strings.Join(dirs, "\n"),
	})
}

// BackupSucceeded writes a BackupSucceeded event.
func (a *API) BackupSucceeded(message string) error {
	return a.WriteEvent(&event.Event{
		Type:      event.BackupSucceeded,
		Timestamp: time.Now(),
		Hostname:  a.hostname,
		Message:   message,
	})
}

// BackupFailed writes a BackupFailed event.
func (a *API) BackupFailed(message string) error {
	return a.WriteEvent(&event.Event{
		Type:      event.BackupFailed,
		Timestamp: time.Now(),
		Hostname:  a.hostname,
		Message:   message,
	})
}
