package mqttClient

import (
	"log"
	"maps"
	"slices"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/pkg/mqtttopic"
)

func (c *ClientStruct) runRouter() {
	var cm *autopaho.ConnectionManager
	routes := make(map[string]map[uint64]MessageHandler) // route -> routeId -> handler
	removed := make(map[uint64]struct{})                 // unsubscribe received before subscribe received

	for {
		select {
		case <-c.ctx.Done():
			c.publishBacklog.Close()
			return

		case cm = <-c.connectedChannel:
			topics := slices.Collect(maps.Keys(routes))
			c.enqueueSubscribe(cm, true, topics...)

		case req := <-c.routeChannel:
			// add new route
			if req.handler != nil {
				if _, ok := removed[req.id]; ok {
					// remove was received before add route, ignore it
					delete(removed, req.id)
					continue
				}

				handlers, exists := routes[req.route]
				if !exists {
					handlers = make(map[uint64]MessageHandler)
					routes[req.route] = handlers
					c.enqueueSubscribe(cm, true, req.route)
				}

				handlers[req.id] = req.handler
				continue
			}

			// remove route
			handlers, ok := routes[req.route]
			if !ok {
				removed[req.id] = struct{}{}
				continue
			}
			if _, ok := handlers[req.id]; !ok {
				removed[req.id] = struct{}{}
				continue
			}

			delete(handlers, req.id)
			if len(handlers) > 0 {
				continue
			}

			// no handler left, unsubscribe
			delete(routes, req.route)
			if cm != nil {
				c.enqueueSubscribe(cm, false, req.route)
			}

		case req := <-c.dispatchChannel:
			for route, handlers := range routes {
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
		}
	}
}

func (c *ClientStruct) onPublishReceived(p paho.PublishReceived) (bool, error) {
	select {
	case <-c.ctx.Done(): // the router has stopped, drop the message
	case c.dispatchChannel <- Message{topic: p.Packet.Topic, payload: p.Packet.Payload}:
	}
	return true, nil
}

func (c *ClientStruct) onConnectionUp(cm *autopaho.ConnectionManager, connack *paho.Connack) {
	log.Printf("mqttClient[%s]: connection is up", c.cfg.Name())

	// hand over to the run routine which (re)subscribes all topics
	select {
	case <-c.ctx.Done():
	case c.connectedChannel <- cm:
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
