package mqttClient

import (
	"strings"
)

const availabilityRetain = true
const availabilityOnline = "online"
const availabilityOffline = "offline"

func (c *ClientStruct) AvailabilityEnabled() bool {
	return len(c.cfg.AvailabilityTopic()) > 0
}

func (c *ClientStruct) GetAvailabilityTopic() string {
	return ReplaceTemplate(c.cfg.AvailabilityTopic(), c.cfg)
}

func ReplaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}
