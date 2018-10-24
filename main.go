package main

import (
	"log"
	"net/http"
	"os"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/julienschmidt/httprouter"
)

var startTime time.Time   //Starttime of the REST server
var session *mgo.Session  //Master connection to the database
var databaseName string   //Name of the database to connect to
var collectionName string //Name of the collection of the database

const pageSize int = 5 //Size of the array of tracks returned by /api/ticker

func main() {
	databaseName = "paragliding_igc" //Name of the database to connect to
	collectionName = "tracks"        //Name of the collection of the database
	var err error
	session, err = mgo.Dial(os.Getenv("DBURL")) //Master connection to the database
	if err != nil {
		log.Fatal("Database-connection could not be made: ", err)
		return
	}
	defer session.Close()
	startTime = time.Now() //Starttime of the REST server
	r := httprouter.New()
	r.GET("/paragliding/api", metaHandler)
	r.POST("/paragliding/api/track", postTrackHandler)
	r.GET("/paragliding/api/track", getTracklistHandler)
	r.GET("/paragliding/api/track/:id", getSingleTrackHandler)
	r.GET("/paragliding/api/track/:id/:field", getSingleTrackFieldHandler)
	//r.GET("/paragliding/api/ticker/latest", getLatestTickerHandler)	//Removed as the router does not handle /latest as well as the wildcard /:timestamp (overlapping routes)
	r.GET("/paragliding/api/ticker", getTickersHandler)
	r.GET("/paragliding/api/ticker/:timestamp", getSpecifiedTickerHandler)
	r.POST("/paragliding/api/webhook/new_track", postNewWebhookHandler)
	r.GET("/paragliding/api/webhook/new_track/:webhookId", getRegisteredWebhookHandler)
	r.DELETE("/paragliding/api/webhook/new_track/:webhookId", deleteRegisteredWebhookHandler)
	r.GET("/admin/api/tracks_count", getTrackCountHandler)
	r.DELETE("/admin/api/tracks", deleteAllTracksHandler)
	err = http.ListenAndServe(":"+os.Getenv("PORT"), r) //Starts the webserver
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
