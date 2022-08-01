package mqttClient

import "time"

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}
