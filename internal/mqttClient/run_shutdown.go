package mqttClient

import (
	"context"
	"log"

	"github.com/eclipse/paho.golang/autopaho"
)

// runShutdown waits for the shutdown to be requested and then closes the client.
func (c *ClientStruct) runShutdown(cm *autopaho.ConnectionManager) {
	defer close(c.done)
	<-c.ctx.Done()
	c.publishBacklog.Close()
	c.disconnect(cm)
}

// disconnect from the broker and closes the client.
func (c *ClientStruct) disconnect(cm *autopaho.ConnectionManager) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// publish availability offline
	if c.cfg.AvailabilityClient().Enabled() && !c.cfg.ReadOnly() {
		msg := c.availabilityMsg(availabilityOffline)
		_, err := cm.Publish(ctx, &msg)
		if err != nil {
			log.Printf("mqttClient[%s]: error publishing offline availability: %s", c.cfg.Name(), err)
		}
	}

	if err := cm.Disconnect(ctx); err != nil {
		log.Printf("mqttClient[%s]: error during disconnect: %s", c.cfg.Name(), err)
	}

	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}
