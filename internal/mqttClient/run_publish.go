package mqttClient

import (
	"errors"
	"log"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func (c *ClientStruct) enqueuePublish(p paho.Publish) {
	if err := c.publishBacklog.Enqueue(p); err != nil {
		log.Printf("mqttClient[%s]: message dropped: topic=%s: %s", c.cfg.Name(), p.Topic, err)
	}
}

// runPublish lives as long as the client; after every (re)connection it drains the backlog.
// A wake-up arriving while a previous drain is still failing stays buffered and is handled right after,
// so a reconnect can never be missed.
func (c *ClientStruct) runPublish() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case cm := <-c.publishWake:
			c.publishOnline(cm)
			c.drainBacklog(cm)
		}
	}
}

// publishOnline bypasses the backlog so the availability is announced first and can never be dropped.
func (c *ClientStruct) publishOnline(cm *autopaho.ConnectionManager) {
	if !c.cfg.AvailabilityClient().Enabled() || c.cfg.ReadOnly() {
		return
	}
	msg := c.availabilityMsg(availabilityOnline)
	if _, err := cm.Publish(c.ctx, &msg); err != nil {
		log.Printf("mqttClient[%s]: error publishing online availability: %s", c.cfg.Name(), err)
	}
}

// drainBacklog forwards the backlog to the broker until the backlog is closed or a publish fails.
func (c *ClientStruct) drainBacklog(cm *autopaho.ConnectionManager) {
	for e := range c.publishBacklog.Dequeue() {
		pr, err := cm.Publish(c.ctx, &e.Value)

		// message rejected by the client library
		if errors.Is(err, paho.ErrInvalidArguments) {
			log.Printf("mqttClient[%s]: invalid message, discarding: topic=%s: %s", c.cfg.Name(), e.Value.Topic, err)
			e.Next(true)
			continue
		}

		// message rejected by the broker
		if pr != nil && pr.ReasonCode >= 0x80 {
			log.Printf("mqttClient[%s]: message rejected by broker, discarding: topic=%s reason=%d: %v",
				c.cfg.Name(), e.Value.Topic, pr.ReasonCode, err,
			)
			e.Next(true)
			continue
		}

		if err != nil {
			// connection problem: keep the message at the front of the backlog and retry after reconnect
			e.Next(false)
			if c.cfg.LogDebug() {
				log.Printf("mqttClient[%s]: error during publish, keeping in backlog: %s", c.cfg.Name(), err)
			}
			return
		}
		e.Next(true)
	}
}
