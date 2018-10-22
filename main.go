package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

var startTime time.Time

func main() {
	os.Setenv("PORT", "8080")
	startTime = time.Now()
	r := httprouter.New()
	r.GET("/paragliding/api", metaHandler)
	r.POST("/paragliding/api/track", postTrackHandler)
	r.GET("/paragliding/api/track", getTracklistHandler)
	r.GET("/paragliding/api/track/:id", getSingleTrackHandler)
	r.GET("/paragliding/api/track/:id/:field", getSingleTrackFieldHandler)
	//r.GET("/paragliding/api/ticker/latest", getLatestTickerHandler)	//Removed as the router does not handle /latest as well as the wildcard /:timestamp
	r.GET("/paragliding/api/ticker", getTickersHandler)
	r.GET("/paragliding/api/ticker/:timestamp", getSpecifiedTickerHandler)
	r.POST("/paragliding/api/webhook/new_track", postNewWebhookHandler)
	r.GET("/paragliding/api/webhook/new_track/:webhookId", getRegisteredWebhookHandler)
	r.DELETE("/paragliding/api/webhook/new_track/:webhookId", deleteRegisteredWebhookHandler)
	r.GET("/admin/api/tracks_count", getTrackCountHandler)
	r.DELETE("/admin/api/tracks", deleteAllTracksHandler)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), r) //Starts the webserver
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}
