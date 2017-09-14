package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/prattmic/restic-remote/auth0"
	"github.com/prattmic/restic-remote/event"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	http.Handle("/api/v1/event", v.ValidateWithScopes([]string{"write:events"}, http.HandlerFunc(writeEvent)))
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!\n")
}

func writeEvent(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST requests allowed")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	var e event.Event
	if err := json.Unmarshal(b, &e); err != nil {
		log.Printf("Failed to decode event %q: %v", string(b), err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Malformed event")
		return
	}

	// Stash the event in datastore.
	key := datastore.NewIncompleteKey(ctx, "Event", nil)
	if _, err := datastore.Put(ctx, key, &e); err != nil {
		log.Printf("Failed to store event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	fmt.Fprintf(w, "Thanks!")
}
