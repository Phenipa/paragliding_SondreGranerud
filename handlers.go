package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/julienschmidt/httprouter"
	igc "github.com/marni/goigc"
)

/*
---------------------------TODO---------------------------
Getting /api/ticker and /api/ticker/<timestamp> are so similar, they should be a function which returns the ticker object
---------------------------TODO---------------------------
*/

type metaData struct { //Output for /paragliding/api
	Version string `json:"name"`
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
}

type jsonTrack struct { //Helper-struct to appropriately respond with data about a requested track.
	Pilot       string        `json:"pilot"`
	Hdate       string        `json:"h_date"`
	Glider      string        `json:"glider"`
	GliderID    string        `json:"glider_id"`
	TrackLength float64       `json:"track_length"`
	URL         string        `json:"url"`
	ID          bson.ObjectId `json:"id" bson:"_id"`
}

type jsonTicker struct {
	TLatest    int64           `json:"t_latest"`
	TStart     int64           `json:"t_start"`
	TStop      int64           `json:"t_stop"`
	Tracks     []bson.ObjectId `json:"tracks"`
	Processing time.Duration   `json:"processing"`
}

type webhook struct {
	URL          string `json:"webhookURL"`
	TriggerValue int64  `json:"minTriggerValue"`
}

func metaHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	meta := metaData{"v1.0", uptime(), "Service for IGC tracks."}   //Create an object which contains response
	json.NewEncoder(w).Encode(meta)                                 //Encode response to json and respond
}

func postTrackHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Header.Get("Content-Type") != "application/json" { //If request is not of type JSON
		http.Error(w, http.StatusText(http.StatusBadRequest)+"\nRequest needs JSON body", http.StatusBadRequest) //Respond that the request needs to be correctly formatted
	}
	var track jsonTrack
	err := json.NewDecoder(r.Body).Decode(&track)
	if err != nil {
		log.Fatal("Decoding of URL failed ", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	trackFile, err := igc.ParseLocation(track.URL)
	if err != nil {
		log.Fatal("Track parsing failed: ", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
	track.ID = bson.NewObjectId()
	postSession := session.Copy()
	defer postSession.Close()
	c := postSession.DB(databaseName).C(collectionName)
	err = c.Insert(track)
	if err != nil {
		log.Fatal("Track could not be inserted: ", err)
	}
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	json.NewEncoder(w).Encode(track.ID)
}

func getTracklistHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	getTracklistSession := session.Copy()
	defer getTracklistSession.Close()
	c := getTracklistSession.DB(databaseName).C(collectionName)
	var indexes []jsonTrack
	err := c.Find(nil).All(&indexes)
	if err != nil {
		log.Fatal("Could not find indexes: ", err)
	}
	ids := make([]interface{}, len(indexes))
	for i := range indexes {
		ids[i] = indexes[i].ID
	}
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	json.NewEncoder(w).Encode(ids)
}

func getSingleTrackHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	getTrackSession := session.Copy()
	defer getTrackSession.Close()
	c := getTrackSession.DB(databaseName).C(collectionName)
	var result jsonTrack
	err := c.FindId(bson.ObjectIdHex(p.ByName("id"))).One(&result)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	json.NewEncoder(w).Encode(result)
}

func getSingleTrackFieldHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	getTrackfieldSession := session.Copy()
	defer getTrackfieldSession.Close()
	c := getTrackfieldSession.DB(databaseName).C(collectionName)
	var result jsonTrack
	err := c.Find(bson.M{"id": p.ByName("id")}).One(&result)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	switch path := p.ByName("field"); path {
	case "pilot":
		fmt.Fprintln(w, result.Pilot)
	case "glider":
		fmt.Fprintln(w, result.Glider)
	case "glider_id":
		fmt.Fprintln(w, result.GliderID)
	case "track_length":
		fmt.Fprintln(w, result.TrackLength)
	case "H_date":
		fmt.Fprintln(w, result.Hdate)
	case "track_src_url":
		fmt.Fprintln(w, result.URL)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/*func getLatestTickerHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {	//Removed as the router does not handle /latest as well as the wildcard /:timestamp (overlapping routes)

}*/

func getTickersHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	processStart := time.Now()
	var tickerResponse jsonTicker
	tickerResponse.Tracks = make([]bson.ObjectId, pageSize, pageSize)
	getTickerListSession := session.Copy()
	defer getTickerListSession.Close()
	c := getTickerListSession.DB(databaseName).C(collectionName)
	result := make([]jsonTrack, pageSize, pageSize)
	var latest jsonTrack
	var start jsonTrack
	dbSize, _ := c.Count()
	err := c.Find(nil).Skip(dbSize - 5).All(&result)
	if err != nil {
		log.Fatal("Could not find entries: ", err)
	}
	err = c.Find(nil).Skip(dbSize - 1).One(&latest)
	if err != nil {
		log.Fatal("Could not find latest entry: ", err)
	}
	err = c.Find(nil).One(&start)
	if err != nil {
		log.Fatal("Could not find first entry: ", err)
	}
	tickerResponse.TLatest = latest.ID.Time().Unix()
	for i, r := range result {
		tickerResponse.Tracks[i] = r.ID
	}
	tickerResponse.TStart = start.ID.Time().Unix()
	tickerResponse.TStop = result[len(result)-1].ID.Time().Unix()
	tickerResponse.Processing = time.Since(processStart) / 1000000  //time.Since returns nanoseconds, dividing it by 1000000 provides the specified unit of milliseconds
	http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
	json.NewEncoder(w).Encode(tickerResponse)
}

func getSpecifiedTickerHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) { //Combination of /api/ticker/latest and /api/ticker/<timestamp> as the router does not handle overlapping routes of static and wildcard-types
	if p.ByName("timestamp") == "latest" { //Handles the path /api/ticker/latest
		getLatestTickerSession := session.Copy()
		defer getLatestTickerSession.Close()
		c := getLatestTickerSession.DB(databaseName).C(collectionName)
		var result jsonTrack
		dbSize, _ := c.Count()
		err := c.Find(nil).Skip(dbSize - 1).One(&result)
		if err != nil {
			log.Fatal("Could not find latest entry: ", err)
		}
		fmt.Fprintln(w, result.ID.Time().Unix())
	} else { //Handles the path /api/ticker/<timestamp>
		processStart := time.Now()
		var tickerResponse jsonTicker
		tickerResponse.Tracks = make([]bson.ObjectId, pageSize, pageSize)
		getTickerListSession := session.Copy()
		defer getTickerListSession.Close()
		c := getTickerListSession.DB(databaseName).C(collectionName)
		result := make([]jsonTrack, pageSize, pageSize)
		var latest jsonTrack
		var start jsonTrack
		dbSize, _ := c.Count()
		unixTime, err := strconv.ParseInt(p.ByName("timestamp"), 10, 64)
		if err != nil {
			log.Fatal("Could not convert timestamp to int: ", err)
		}
		tm := time.Unix(unixTime, 0)
		oidtime := bson.NewObjectIdWithTime(tm)
		err = c.Find(bson.M{"_id": bson.M{"$gte": oidtime}}).Limit(pageSize).All(&result)
		if err != nil {
			log.Fatal("Could not find entries: ", err)
		}
		err = c.Find(nil).Skip(dbSize - 1).One(&latest)
		if err != nil {
			log.Fatal("Could not find latest entry: ", err)
		}
		err = c.Find(nil).One(&start)
		if err != nil {
			log.Fatal("Could not find first entry: ", err)
		}
		tickerResponse.TLatest = latest.ID.Time().Unix()
		for i, r := range result {
			tickerResponse.Tracks[i] = r.ID
		}
		tickerResponse.TStart = start.ID.Time().Unix()
		tickerResponse.TStop = result[len(result)-1].ID.Time().Unix()
		tickerResponse.Processing = time.Since(processStart) / 1000000  //time.Since returns nanoseconds, dividing it by 1000000 provides the specified unit of milliseconds
		http.Header.Add(w.Header(), "content-type", "application/json") //Set response-header to json reflect that response is json-formatted
		json.NewEncoder(w).Encode(tickerResponse)
	}
}

func postNewWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Header.Get("Content-Type") != "application/json" { //If request is not of type JSON
		http.Error(w, http.StatusText(http.StatusBadRequest)+"\nRequest needs JSON body", http.StatusBadRequest) //Respond that the request needs to be correctly formatted
	}
	valid := false
	var newWebhook webhook
	err := json.NewDecoder(r.Body).Decode(&newWebhook)
	if err != nil {
		log.Fatal("Decoding of URL failed ", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if strings.Contains(newWebhook.URL, "hooks.slack.com") {
		valid = true
	}
	if strings.Contains(newWebhook.URL, "discordapp.com") {
		newWebhook.URL = newWebhook.URL + "/slack"
		valid = true
	}
	if valid {
		webhookSession := session.Copy()
		defer webhookSession.Close()
		c := webhookSession.DB(databaseName).C(webhookCollection)
		err = c.Insert(newWebhook)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			log.Fatal("Webhook could not be inserted: ", err)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

func getRegisteredWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var result webhook
	webhookSession := session.Copy()
	defer webhookSession.Close()
	c := webhookSession.DB(databaseName).C(webhookCollection)
	c.Find(bson.M{"webhookURL": p.ByName("webhookId")}).One(&result)
	json.NewEncoder(w).Encode(result)
}

func deleteRegisteredWebhookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func getTrackCountHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}

func deleteAllTracksHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}
