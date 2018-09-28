package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/marni/goigc"
)

type metaData struct {
	Version string `json:"name"`
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
}

type urlType struct {
	URL string `json:"url"`
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
		if r.Body == nil {
			http.Error(w, "Request needs JSON body", http.StatusBadRequest)
		} else {
			var url urlType
			err := json.NewDecoder(r.Body).Decode(&url)
			track, err := igc.ParseLocation(url.URL)
			if err == nil {
				id := track.Pilot + track.Date.String()
				_, ok := database[id]
				if !ok {
					database[id] = track
					fmt.Fprint(w, id)
				}
			} else {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			}
		}
	} else if r.Method == "GET" {
		keys := make([]string, 0, len(database))
		for k := range database {
			keys = append(keys, k)
		}
		json.NewEncoder(w).Encode(keys)
	} else {
		http.Error(w, "Not implemented yet", http.StatusNotImplemented)
	}
	return
}

var database map[string]igc.Track

func main() {
	database = make(map[string]igc.Track)
	http.HandleFunc("/igcinfo/api/", handlerRoot)
	http.HandleFunc("/igcinfo/api/igc", handlerIndex)
	http.ListenAndServe(":8080", nil)
}
