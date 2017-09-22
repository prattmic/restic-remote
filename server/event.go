package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prattmic/restic-remote/event"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

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
