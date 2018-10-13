package main

import (
	"net/http"
	"os"
	"time"

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

func init() {
	startTime = time.Now()
}

var database map[string]igc.Track
var startTime time.Time

func main() {
	database = make(map[string]igc.Track)
	http.HandleFunc("/igcinfo/api/", handlerRoot)
	http.HandleFunc("/igcinfo/api/igc/", handlerIndex)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
