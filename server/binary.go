package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prattmic/restic-remote/api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type binaryEntity struct {
	// Timestamp is the time that the entity was created.
	Timestamp time.Time

	// Binary is the actual binary info.
	api.Binary
}

func binary(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only GET requests allowed")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "name parameter required")
		return
	}

	var be []binaryEntity
	q := datastore.NewQuery("Binary").Filter("Name =", name).Order("-Timestamp").Limit(1)
	if _, err := q.GetAll(ctx, &be); err != nil {
		log.Printf("Failed to get binary info for query %+v: %v", q, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	if len(be) < 1 {
		log.Printf("No results for name %s", name)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "name %s not found", name)
		return
	}

	if err := json.NewEncoder(w).Encode(be[0].Binary); err != nil {
		log.Printf("Failed to encode binary info %+v: %v", be[0].Binary, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}
}
