package mqttClient

import (
	"log"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// enqueueBrokerOp never blocks; the operation is executed later by runBrokerOps in enqueuePublish order.
func (c *ClientStruct) enqueueBrokerOp(cm *autopaho.ConnectionManager, subscribe bool, topics ...string) {
	if cm == nil || len(topics) == 0 {
		return
	}
	c.brokerOps.PushBack(brokerOp{cm: cm, subscribe: subscribe, topics: topics})

	select {
	case c.brokerWake <- struct{}{}:
	default: // a pending wake-up is enough
	}
}

// runBrokerOps executes the broker operations one at a time.
// The calls block until acknowledged, which needs the router to keep delivering, so run must never do them inline.
func (c *ClientStruct) runBrokerOps() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.brokerWake:
		}

		for {
			e := c.brokerOps.Front()
			if e == nil {
				break
			}
			op := c.brokerOps.Remove(e)
			if c.ctx.Err() != nil {
				return
			}
			c.execBrokerOp(op)
		}
	}
}

func (c *ClientStruct) execBrokerOp(op brokerOp) {
	if !op.subscribe {
		if _, err := op.cm.Unsubscribe(c.ctx, &paho.Unsubscribe{Topics: op.topics}); err != nil {
			log.Printf("mqttClient[%s]: failed to unsubscribe: %s", c.cfg.Name(), err)
		}
		return
	}

	opts := make([]paho.SubscribeOptions, 0, len(op.topics))
	for _, t := range op.topics {
		opts = append(opts, paho.SubscribeOptions{Topic: t, QoS: byte(1)})
	}
	if _, err := op.cm.Subscribe(c.ctx, &paho.Subscribe{Subscriptions: opts}); err != nil {
		log.Printf("mqttClient[%s]: failed to subscribe: %s", c.cfg.Name(), err)
	}
}
