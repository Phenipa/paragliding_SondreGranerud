package main

import (
	"strconv"
	"time"
)

func uptime() string {
	var years, months, weeks, days, hours, minutes int
	seconds := int(time.Since(startTime).Seconds())
	if seconds >= 60 {
		minutes = int(seconds / 60)
		seconds = seconds % 60
		if minutes >= 60 {
			hours = int(minutes / 60)
			minutes = minutes % 60
			if hours >= 24 {
				days = int(hours / 24)
				hours = hours % 24
				if days >= 7 {
					weeks = int(days / 7)
					days = days % 7
					if weeks >= 4 {
						months = int(weeks / 4)
						weeks = weeks % 4
						if months >= 12 {
							years = int(months / 12)
							months = months % 12
						}
					}
				}
			}
		}
	}
	output := "P" + strconv.Itoa(years) + "Y" + strconv.Itoa(months) + "M" + strconv.Itoa(weeks) + "W" + strconv.Itoa(days) + "DT" + strconv.Itoa(hours) + "H" + strconv.Itoa(minutes) + "M" + strconv.Itoa(seconds) + "S"
	return output
}
