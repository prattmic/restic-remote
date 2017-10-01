package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prattmic/restic-remote/auth0"
)

var (
	auth0JWKS     = os.Getenv("AUTH0_API_JWKS")
	auth0Issuer   = os.Getenv("AUTH0_API_ISSUER")
	auth0Audience = os.Getenv("AUTH0_API_AUDIENCE")
)

func init() {
	if auth0JWKS == "" {
		panic("AUTH0_API_JWKS must be set")
	}
	if auth0Issuer == "" {
		panic("AUTH0_API_ISSUER must be set")
	}
	if auth0Audience == "" {
		panic("AUTH0_API_AUDIENCE must be set")
	}

	v := auth0.NewValidator(auth0JWKS, auth0Issuer, []string{auth0Audience})

	http.Handle("/", v.ValidateWithScopes(nil, http.HandlerFunc(root)))

	releaseScopes := auth0.MethodScopes{
		"GET":  []string{"read:release"},
		"POST": []string{"write:release"},
	}
	http.Handle("/api/v1/release", v.ValidateWithScopes(releaseScopes, http.HandlerFunc(release)))

	eventScopes := auth0.MethodScopes{
		"POST": []string{"write:events"},
	}
	http.Handle("/api/v1/event", v.ValidateWithScopes(eventScopes, http.HandlerFunc(writeEvent)))
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!\n")
}
