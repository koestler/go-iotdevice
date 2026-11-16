package mqttClient

import (
	"log"
	"maps"
	"slices"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/pkg/mqtttopic"
)

func (c *ClientStruct) run() {
	name := c.cfg.Name()
	logDebug := c.cfg.LogDebug()

	var cm *autopaho.ConnectionManager
	subscriptions := make(map[string]map[uint64]MessageHandler) // route -> routeId -> handler
	unsubscribed := make(map[uint64]struct{})                   // unsubscribe received before subscribe received

	for {
		select {
		case <-c.ctx.Done():
			c.publishBacklog.Close()
			return

		case cm = <-c.connectedChannel:
			topics := slices.Collect(maps.Keys(subscriptions))
			c.enqueueBrokerOp(cm, true, topics...)

		case req := <-c.subscriptionChannel:
			if req.handler != nil {
				if _, ok := unsubscribed[req.id]; ok {
					// unsubscribe was received before subscribe, ignore this subscribe
					delete(unsubscribed, req.id)
					continue
				}

				handlers, exists := subscriptions[req.route]
				if !exists {
					handlers = make(map[uint64]MessageHandler)
					subscriptions[req.route] = handlers

					if logDebug {
						log.Printf("mqttClient[%s]: subscribing to topic: %s", name, req.route)
					}
					c.enqueueBrokerOp(cm, true, req.route)
				}

				handlers[req.id] = req.handler
				continue
			}

			handlers, ok := subscriptions[req.route]
			if !ok {
				unsubscribed[req.id] = struct{}{}
				continue
			}
			if _, ok := handlers[req.id]; !ok {
				unsubscribed[req.id] = struct{}{}
				continue
			}

			delete(handlers, req.id)
			if len(handlers) > 0 {
				continue
			}

			// no handler left, unsubscribe
			delete(subscriptions, req.route)
			// c.router.UnregisterHandler(req.route)
			if cm != nil {
				if logDebug {
					log.Printf("mqttClient[%s]: unsubscribing from topic: %s", name, req.route)
				}
				c.enqueueBrokerOp(cm, false, req.route)
			}

		case req := <-c.dispatchChannel:
			for route, handlers := range subscriptions {
				if !mqtttopic.Match(route, req.topic) {
					continue
				}

				for _, h := range handlers {
					h(req)
				}
			}

		case done := <-c.shutdownChannel:
			go func() {
				defer close(done)
				c.shutdown(cm)
			}()
			// keep serving until the context is cancelled by shutdown
		}
	}
}

func (c *ClientStruct) onPublishReceived(p paho.PublishReceived) (bool, error) {
	c.dispatchChannel <- Message{topic: p.Packet.Topic, payload: p.Packet.Payload}
	return true, nil
}

func (c *ClientStruct) onConnectionUp(cm *autopaho.ConnectionManager, connack *paho.Connack) {
	log.Printf("mqttClient[%s]: connection is up", c.cfg.Name())

	// hand over to the run routine which (re)subscribes all topics
	select {
	case <-c.ctx.Done():
	case c.connectedChannel <- cm:
	}

	if c.cfg.AvailabilityClient().Enabled() {
		c.enqueuePublish(c.availabilityMsg(availabilityOnline))
	}

	// wake the publish routine; a pending wake-up is enough, so never block
	select {
	case c.publishWake <- cm:
	default:
	}
}

func (c *ClientStruct) availabilityMsg(payload string) paho.Publish {
	mCfg := c.cfg.AvailabilityClient()

	return paho.Publish{
		QoS:     mCfg.Qos(),
		Topic:   c.cfg.AvailabilityClientTopic(),
		Payload: []byte(payload),
		Retain:  mCfg.Retain(),
	}
}
