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

type releaseEntity struct {
	// Timestamp is the time that the entity was created.
	Timestamp time.Time

	// Release is the actual release info.
	api.Release
}

func release(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		releaseGet(w, r)
	case "POST":
		releasePost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "%s requests not allowed", r.Method)
	}
}

func releaseGet(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var be []releaseEntity
	q := datastore.NewQuery("Release").Order("-Timestamp").Limit(1)
	if _, err := q.GetAll(ctx, &be); err != nil {
		log.Printf("Failed to get release for query %+v: %v", q, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	if len(be) < 1 {
		log.Printf("No release results")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no release found")
		return
	}

	if err := json.NewEncoder(w).Encode(be[0].Release); err != nil {
		log.Printf("Failed to encode release %+v: %v", be[0].Release, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}
}

func releasePost(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	var rel api.Release
	if err := json.Unmarshal(b, &rel); err != nil {
		log.Printf("Failed to decode release %q: %v", string(b), err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Malformed release")
		return
	}

	if rel.Path == "" {
		log.Printf("Release %+v missing path", rel)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "path required")
		return
	}

	if rel.ResticVersion == "" {
		log.Printf("Release %+v missing restic version", rel)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "restic version required")
		return
	}

	if rel.ClientVersion == "" {
		log.Printf("Release %+v missing client version", rel)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "client version required")
		return
	}

	entity := releaseEntity{
		Timestamp: time.Now(),
		Release:   rel,
	}

	// Stash the release in datastore.
	key := datastore.NewIncompleteKey(ctx, "Release", nil)
	if _, err := datastore.Put(ctx, key, &entity); err != nil {
		log.Printf("Failed to store release: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}
}
