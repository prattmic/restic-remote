package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	switch r.Method {
	case "GET":
		binaryGet(w, r)
	case "POST":
		binaryGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "%s requests not allowed", r.Method)
	}
}

func binaryGet(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

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

func binaryPost(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	var bin api.Binary
	if err := json.Unmarshal(b, &bin); err != nil {
		log.Printf("Failed to decode binary %q: %v", string(b), err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Malformed binary")
		return
	}

	if bin.Name == "" {
		log.Printf("Binary %+v missing name", bin)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "name required")
		return
	}

	if bin.Path == "" {
		log.Printf("Binary %+v missing path", bin)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "path required")
		return
	}

	if bin.Version == "" {
		log.Printf("Binary %+v missing version", bin)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "version required")
		return
	}

	entity := binaryEntity{
		Timestamp: time.Now(),
		Binary:    bin,
	}

	// Stash the binary in datastore.
	key := datastore.NewIncompleteKey(ctx, "Binary", nil)
	if _, err := datastore.Put(ctx, key, &entity); err != nil {
		log.Printf("Failed to store binary: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}
}
