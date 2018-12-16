package mqttClient

import "time"

const timeFormat string = "2006-01-02T15:04:05"

func timeToString(t time.Time) (string) {
	return t.UTC().Format(timeFormat)
}
