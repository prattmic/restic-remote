// +build appengine

package auth0

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// requestContext returns the Context for a request.
func requestContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

// httpClient returns a new http.Client.
func httpClient(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}
