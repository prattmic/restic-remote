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

	"github.com/prattmic/restic-remote/auth0"
	"github.com/prattmic/restic-remote/event"
)

// API endpoints.
const (
	eventEndpoint = "/api/v1/event"
)

type Config struct {
	// ClientConfig contains the authentication configuration.
	auth0.ClientConfig `mapstructure:",squash"`

	// Root is the API root URL.
	Root string `mapstructure:"root"`
}

// API describes a configured API target.
type API struct {
	// root is the API root URL.
	//
	// root is immutable.
	root url.URL

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
		root:   *u,
		client: client,
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

// WriteEvent sends an event to the server.
func (a *API) WriteEvent(e *event.Event) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(e); err != nil {
		return fmt.Errorf("error encoding event %+v: %v", e, err)
	}

	u := a.url(eventEndpoint)
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

	return fmt.Errorf("error response when writing event: %+v\n%s", r, string(b))
}
