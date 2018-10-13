package main

import (
	"net/http"
	"os"
	"time"

	"github.com/marni/goigc"
)

type metaData struct { //Output for /igcinfo/api
	Version string `json:"name"`
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
}

type urlType struct { //Helper-struct to appropriately extract an IGC-url from a POST-request
	URL string `json:"url"`
}

type idType struct { //Helper-struct to appropriately respond with the id of a particular track after being posted
	ID string `json:"id"`
}

type jsonTrack struct { //Helper-struct to appropriately respond with data about a requested track.
	Pilot       string  `json:"pilot"`
	Hdate       string  `json:"h_date"`
	Glider      string  `json:"glider"`
	GliderID    string  `json:"glider_id"`
	TrackLength float64 `json:"track_length"`
}

var database map[string]igc.Track //-Runs on first startup to set global variables appropriately
var startTime time.Time           ///

func main() {
	startTime = time.Now()
	database = make(map[string]igc.Track)              //Initializes database
	http.HandleFunc("/igcinfo/api/", handlerRoot)      //Handles /igcinfo/api/
	http.HandleFunc("/igcinfo/api/igc/", handlerIndex) //Handles /igcinfo/api/igc/ and its sub-paths
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)    //Starts the webserver
}
