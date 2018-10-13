package main

import "github.com/segmentio/ksuid"

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
