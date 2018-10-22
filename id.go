package main

import (
	"log"
	"os"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/segmentio/ksuid"
)

func genUniqueID() string { //This function generates a unique id for all new POST-requests. It is recursive to ensure that if an existing id is generated again it won't use it, but rather generate a new one
	session, err := mgo.Dial(os.Getenv("DBURL"))
	if err != nil {
		log.Fatal("Database.connection could not be made: ", err)
		return ""
	}
	c := session.DB("paragliding_igc").C("tracks")
	id := ksuid.New() //Generate an id
	var result jsonTrack
	c.Find(bson.M{"id": id.String()}).One(&result)
	if result.ID == "" { //Check that the id does not alread exist in the database
		return id.String() //If not in database, return id
	}
	return genUniqueID() //If already in databse, run once more to generate a new id
}
