package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	igc "github.com/marni/goigc"
)

func handlerRoot(w http.ResponseWriter, r *http.Request) { //Handles /igcinfo/api/
	if parts := strings.Split(r.URL.Path, "/"); parts[3] != "" { //Check that the url is valid
		http.Error(w, "404 Not found", http.StatusNotFound)
	} else {
		http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
		meta := metaData{"v1.0", uptime(), "Service for IGC tracks."}   //Create an object which contains response
		json.NewEncoder(w).Encode(meta)                                 //Encode response to json and respond
	}
}

func handlerIndex(w http.ResponseWriter, r *http.Request) { //Handles /igcinfo/api/igc/
	if r.Method == "POST" { //If request is POST
		if r.Header.Get("Content-Type") != "application/json" { //If request is not of type JSON
			http.Error(w, http.StatusText(http.StatusBadRequest)+"\nRequest needs JSON body", http.StatusBadRequest) //Respond that the request needs to be correctly formatted
		} else {
			var url urlType                             //Initiate a url-helper-type
			err := json.NewDecoder(r.Body).Decode(&url) //Decode content of request (the url which was posted)
			track, err := igc.ParseLocation(url.URL)    //Use goigc-library to extract a track from the url that was provided
			if err == nil {                             //If the url is deemed valid by goigc
				id := genUniqueID()                 //Generate an id for this track
				if _, here := database[id]; !here { //Final check to ensure id is not already in database (this should be unnecessary as the genUniqueId function already makes this check)
					//This does not ensure that the track is not already in the database. The ids I generate make it such that I would need to parse the whole database
					//and ensure all contents of existing tracks do not match the track which is being added. I deemed it unnecessary to do this, even though I risk
					//filling the database with exclusively repeat-data.
					database[id] = track                                            //Add the track to the database
					ids := idType{id}                                               //Put the id into a struct to ready it for response
					http.Header.Add(w.Header(), "content-type", "application/json") //Set response header to reflect the content being JSON
					json.NewEncoder(w).Encode(ids)                                  //Generate response (the id which was given to the track that was POSTed)
				}
			} else {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest) //Handles bad requests
			}
		}
	} else if r.Method == "GET" { //If request is GET
		if parts := strings.Split(r.URL.Path, "/"); len(parts[4]) == 0 { //Checks the incoming URL-path to route the request appropriately
			keys := make([]string, 0, len(database)) //Ready an array to hold all ids that are present in the database.
			for k := range database {                //Parse the database and extract the keys
				keys = append(keys, k) //Put the keys into the array
			}
			json.NewEncoder(w).Encode(keys) //Encode the response into a JSON-array and respond
		} else { //If the GET-request is /igcinfo/api/<id>/ or its sub-paths
			if foundTrack, here := database[parts[4]]; here { //If the id exists in the database
				if parts[5] != "" { //If the GET-request is /igcinfo/api/<id>/<field>
					if parts[5] == "pilot" { //Handle the <field> path and respond accordingly
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
					} else { //Requested path was not found i.e. /igcinfo/api/igc/<valid id>/<gibberish>
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
				} else if parts[5] == "" || len(parts) == 5 { //If the requested path is exactly /igcinfo/api/igc/<id>
					jsontrack := jsonTrack{Pilot: foundTrack.Pilot, Hdate: foundTrack.Date.String(), Glider: foundTrack.GliderType, GliderID: foundTrack.GliderID, TrackLength: foundTrack.Task.Distance()}
					json.NewEncoder(w).Encode(jsontrack) //Format all information and respond
				}
			} else {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound) //Requested path is invalid
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented) //Request-type is not handled. This is the case if r.Method is not either GET or POST
	}
}
