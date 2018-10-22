package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/globalsign/mgo"
	"github.com/julienschmidt/httprouter"
	igc "github.com/marni/goigc"
	"github.com/globalsign/mgo/bson"
)

type metaData struct { //Output for /paragliding/api
	Version string `json:"name"`
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
}

type jsonTrack struct { //Helper-struct to appropriately respond with data about a requested track.
	Pilot       string  `json:"pilot"`
	Hdate       string  `json:"h_date"`
	Glider      string  `json:"glider"`
	GliderID    string  `json:"glider_id"`
	TrackLength float64 `json:"track_length"`
	URL         string  `json:"url"`
	_ID 		bson.ObjectId `json:"id"`
}

func metaHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	meta := metaData{"v1.0", uptime(), "Service for IGC tracks."}   //Create an object which contains response
	json.NewEncoder(w).Encode(meta)                                 //Encode response to json and respond
}

func postTrackHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Header.Get("Content-Type") != "application/json" { //If request is not of type JSON
		http.Error(w, http.StatusText(http.StatusBadRequest)+"\nRequest needs JSON body", http.StatusBadRequest) //Respond that the request needs to be correctly formatted
		return
	}
	var track jsonTrack
	err := json.NewDecoder(r.Body).Decode(&track)
	if err != nil {
		log.Fatal("Decoding of URL failed ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	trackFile, err := igc.ParseLocation(track.URL)
	if err != nil {
		log.Fatal("Track parsing failed: ", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	track.Pilot = trackFile.Pilot
	track.Hdate = trackFile.Date.String()
	track.Glider = trackFile.GliderType
	track.GliderID = trackFile.GliderID
	var length float64
	for i := range trackFile.Points {
		if i < len(trackFile.Points)-1 {
			length += trackFile.Points[i].Distance(trackFile.Points[i+1])
		}
	}
	track.TrackLength = length
	track._ID = bson.NewObjectId()
	os.Setenv("DBURL", "mongodb://access:paragl1ding_access@ds237363.mlab.com:37363/paragliding_igc")
	session, err := mgo.Dial(os.Getenv("DBURL"))
	if err != nil {
		log.Fatal("Database-connection could not be made: ", err)
	}
	c := session.DB("paragliding_igc").C("tracks")
	err = c.Insert(track)
	if err!= nil {
		log.Fatal("Track could not be inserted: ", err)
		return
	}
	session.Close()
	json.NewEncoder(w).Encode(track._ID)
}

func getTracklistHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	os.Setenv("DBURL", "mongodb://access:paragl1ding_access@ds237363.mlab.com:37363/paragliding_igc")
	session, err := mgo.Dial(os.Getenv("DBURL"))
	if err != nil {
		log.Fatal("Database-connection could not be made: ", err)
		return
	}
	c := session.DB("paragliding_igc").C("tracks")
	var indexes []jsonTrack
	err = c.Find(nil).All(&indexes)
	if err != nil {
		log.Fatal("Could not find indexes: ", err)
		return
	}
	session.Close()
	json.NewEncoder(w).Encode(indexes)
}

func getSingleTrackHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func getSingleTrackFieldHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

/*func getLatestTickerHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}*/

func getTickersHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func getSpecifiedTickerHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func postNewWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func getRegisteredWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func deleteRegisteredWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func getTrackCountHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func deleteAllTracksHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}
