package mqttClient

import (
	"log"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// enqueueSubscribe never blocks; the operation is executed later by runSubscribe in enqueuePublish order.
func (c *ClientStruct) enqueueSubscribe(cm *autopaho.ConnectionManager, subscribe bool, topics ...string) {
	if cm == nil || len(topics) == 0 {
		return
	}
	c.subscribeBacklog.PushBack(brokerOp{cm: cm, subscribe: subscribe, topics: topics})

	select {
	case c.subscribeWake <- struct{}{}:
	default: // a pending wake-up is enough
	}
}

// runSubscribe executes the broker operations one at a time.
// The calls block until acknowledged, which needs the router to keep delivering, so runRouter must never do them inline.
func (c *ClientStruct) runSubscribe() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.subscribeWake:
		}

		for {
			e := c.subscribeBacklog.Front()
			if e == nil {
				break
			}
			op := c.subscribeBacklog.Remove(e)
			if c.ctx.Err() != nil {
				return
			}
			c.execSubscribe(op)
		}
	}
}

func (c *ClientStruct) execSubscribe(op brokerOp) {
	if !op.subscribe {
		log.Printf("mqttClient[%s]: unsubscribing from topics: %v", c.cfg.Name(), op.topics)

		if _, err := op.cm.Unsubscribe(c.ctx, &paho.Unsubscribe{Topics: op.topics}); err != nil {
			log.Printf("mqttClient[%s]: failed to unsubscribe: %s", c.cfg.Name(), err)
		}
		return
	}

	log.Printf("mqttClient[%s]: subscribing to topics: %v", c.cfg.Name(), op.topics)

	opts := make([]paho.SubscribeOptions, 0, len(op.topics))
	for _, t := range op.topics {
		opts = append(opts, paho.SubscribeOptions{Topic: t, QoS: byte(1)})
	}
	if _, err := op.cm.Subscribe(c.ctx, &paho.Subscribe{Subscriptions: opts}); err != nil {
		log.Printf("mqttClient[%s]: failed to subscribe: %s", c.cfg.Name(), err)
	}
}
