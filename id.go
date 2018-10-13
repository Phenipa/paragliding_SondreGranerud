package main

import "github.com/segmentio/ksuid"

func genUniqueID() string { //This function generates a unique id for all new POST-requests. It is recursive to ensure that if an existing id is generated again it won't use it, but rather generate a new one
	id := ksuid.New()       //Generate an id
	if len(database) == 0 { //If this is the first entry into the database, simply allow it
		return id.String()
	}
	if _, here := database[id.String()]; !here { //Check that the id does not alread exist in the database
		return id.String() //If not in database, return id
	}
	return genUniqueID() //If already in databse, run once more to generate a new id
}
