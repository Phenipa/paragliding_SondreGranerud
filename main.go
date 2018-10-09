package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"

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

type jsonTrack struct {
	Pilot       string  `json:"pilot"`
	Hdate       string  `json:"h_date"`
	Glider      string  `json:"glider"`
	GliderID    string  `json:"glider_id"`
	TrackLength float64 `json:"track_length"`
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

func uptime() string {
	var years int
	var months int
	var weeks int
	var days int
	var hours int
	var minutes int
	seconds := int(time.Since(startTime).Seconds())
	if seconds >= 60 {
		minutes = int(seconds / 60)
		seconds = seconds % 60
		if minutes >= 60 {
			hours = int(minutes / 60)
			minutes = minutes % 60
			if hours >= 24 {
				days = int(hours / 24)
				hours = hours % 24
				if days >= 7 {
					weeks = int(days / 7)
					days = days % 7
					if weeks >= 4 {
						months = int(weeks / 4)
						weeks = weeks % 4
						if months >= 12 {
							years = int(months / 12)
							months = months % 12
						}
					}
				}
			}
		}
	}
	output := "P" + strconv.Itoa(years) + "Y" + strconv.Itoa(months) + "M" + strconv.Itoa(weeks) + "W" + strconv.Itoa(days) + "DT" + strconv.Itoa(hours) + "H" + strconv.Itoa(minutes) + "M" + strconv.Itoa(seconds) + "S"
	return output
}

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
				http.Error(w, http.StatusText(http.StatusBadRequest)+"\n"+err.Error(), http.StatusBadRequest)
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
				} else if (len(parts) > 7 && parts[5] == "") || len(parts) == 5 {
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

func init() {
	startTime = time.Now()
}

var database map[string]igc.Track
var startTime time.Time

func main() {
	database = make(map[string]igc.Track)
	http.HandleFunc("/igcinfo/api/", handlerRoot)
	http.HandleFunc("/igcinfo/api/igc/", handlerIndex)
	http.ListenAndServe(":8080", nil)
	appengine.Main()
}
