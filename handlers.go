package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	igc "github.com/marni/goigc"
)

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	if parts := strings.Split(r.URL.Path, "/"); parts[3] != "" {
		http.Error(w, "404 Not found", http.StatusNotFound)
	} else {
		http.Header.Add(w.Header(), "content-type", "application/json")
		meta := metaData{"v1.0", uptime(), "Service for IGC tracks."}
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
					http.Header.Add(w.Header(), "content-type", "application/json")
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
						var length float64
						for i := range foundTrack.Points {
							if i < len(foundTrack.Points)-1 {
								length += foundTrack.Points[i].Distance(foundTrack.Points[i+1])
							}
						}
						fmt.Fprintln(w, length)
					} else {
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
				} else if parts[5] == "" || len(parts) == 5 {
					jsontrack := jsonTrack{Pilot: foundTrack.Pilot, Hdate: foundTrack.Date.String(), Glider: foundTrack.GliderType, GliderID: foundTrack.GliderID, TrackLength: foundTrack.Task.Distance()}
					json.NewEncoder(w).Encode(jsontrack)
				}
			} else {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}
