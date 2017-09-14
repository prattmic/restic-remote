// +build !appengine

package auth0

import (
	"context"
	"net/http"
)

// requestContext returns the Context for a request.
func requestContext(r *http.Request) context.Context {
	return r.Context()
}

// httpClient returns a new http.Client.
func httpClient(context.Context) *http.Client {
	return http.DefaultClient
}
