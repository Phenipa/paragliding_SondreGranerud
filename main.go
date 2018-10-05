package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/marni/goigc"

	"github.com/segmentio/ksuid"
)

type metaData struct {
	Version string `json:"name"`
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
}

type urlType struct {
	URL string `json:"url"`
}

type idType struct {
	ID string `json:"id"`
}

func genUniqueID() string {
	id := ksuid.New()
	if len(database) == 0 {
		return id.String()
	}
	for range database {
		if _, here := database[id.String()]; !here {
			return id.String()
		}
	}
	return genUniqueID()
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	if parts := strings.Split(r.URL.Path, "/"); parts[3] != "" {
		http.Error(w, "404 Not found", http.StatusNotFound)
	} else {
		http.Header.Add(w.Header(), "content-type", "application/json")
		meta := metaData{"v1.0", "Placeholder uptime", "Service for IGC tracks."}
		json.NewEncoder(w).Encode(meta)
	}
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusBadRequest)+"\nRequest needs JSON body", http.StatusBadRequest)
		} else {
			var url urlType
			err := json.NewDecoder(r.Body).Decode(&url)
			track, err := igc.ParseLocation(url.URL)
			if err == nil {
				id := genUniqueID()
				if _, here := database[id]; !here {
					database[id] = track
					ids := idType{id}
					json.NewEncoder(w).Encode(ids)
				}
			} else {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			}
		}
	} else if r.Method == "GET" {
		if parts := strings.Split(r.URL.Path, "/"); len(parts[4]) == 0 {
			keys := make([]string, 0, len(database))
			for k := range database {
				keys = append(keys, k)
			}
			json.NewEncoder(w).Encode(keys)
		} else {
			if foundTrack, here := database[parts[4]]; here {
				if parts[5] != "" {
					if parts[5] == "pilot" {
						fmt.Fprintln(w, foundTrack.Pilot)
					} else if parts[5] == "h_date" {
						fmt.Fprintln(w, foundTrack.Date.String())
					} else if parts[5] == "glider" {
						fmt.Fprintln(w, foundTrack.GliderType)
					} else if parts[5] == "glider_id" {
						fmt.Fprintln(w, foundTrack.GliderID)
					} else if parts[5] == "track_length" {
						fmt.Fprintln(w, foundTrack.Task.Distance())
					} else {
						http.Error(w, http.StatusText(http.StatusBadRequest)+"\nUnknown field.", http.StatusBadRequest)
					}
				}
			} else {
				http.Error(w, http.StatusText(http.StatusBadRequest)+"\nCould not find that track-id", http.StatusBadRequest)
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

var database map[string]igc.Track

func main() {
	database = make(map[string]igc.Track)
	http.HandleFunc("/igcinfo/api/", handlerRoot)
	http.HandleFunc("/igcinfo/api/igc/", handlerIndex)
	http.ListenAndServe(":8080", nil)
}
