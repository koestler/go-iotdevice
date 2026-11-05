package mqttClient

import (
	"context"
	"log"
	"maps"
	"slices"
	"sync"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/internal/queue"
)

type ClientStruct struct {
	cfg Config

	cliCfg         autopaho.ClientConfig
	cm             *autopaho.ConnectionManager
	router         *paho.StandardRouter
	publishBacklog *queue.Fifo[*paho.Publish]

	subscriptionsMutex sync.RWMutex
	subscriptions      map[string]map[uint64]MessageHandler // subscribeTopic -> routeId -> handler
	nextRouteId        uint64

	ctx    context.Context
	cancel context.CancelFunc
}

func (c *ClientStruct) Name() string {
	return c.cfg.Name()
}

func (c *ClientStruct) GetCtx() context.Context {
	return c.ctx
}

// AddRoute registers a handler for subscribeTopic and subscribes on the broker if needed.
// The returned function removes the handler again; the topic is unsubscribed once its last handler is removed.
func (c *ClientStruct) AddRoute(subscribeTopic string, handler MessageHandler) (remove func()) {
	if c.cfg.LogMessages() {
		handler = logWrapper(c.cfg.Name(), handler)
	}

	c.subscriptionsMutex.Lock()
	defer c.subscriptionsMutex.Unlock()

	handlers, exists := c.subscriptions[subscribeTopic]
	if !exists {
		handlers = make(map[uint64]MessageHandler)
		c.subscriptions[subscribeTopic] = handlers

		// the router has one dispatcher per topic; it forwards to all handlers currently registered
		c.router.RegisterHandler(subscribeTopic, func(p *paho.Publish) {
			c.dispatch(subscribeTopic, Message{topic: p.Topic, payload: p.Payload})
		})
	}

	routeId := c.nextRouteId
	c.nextRouteId++
	handlers[routeId] = handler

	if !exists && c.cm != nil {
		_, _ = c.cm.Subscribe(c.ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{subscribeOptions(subscribeTopic)},
		})
	}

	return func() {
		c.subscriptionsMutex.Lock()
		defer c.subscriptionsMutex.Unlock()

		handlers, ok := c.subscriptions[subscribeTopic]
		if !ok {
			return
		}
		delete(handlers, routeId)
		if len(handlers) > 0 {
			return
		}

		// no handler left, unsubscribe
		delete(c.subscriptions, subscribeTopic)
		c.router.UnregisterHandler(subscribeTopic)
		if c.cm != nil {
			_, _ = c.cm.Unsubscribe(c.ctx, &paho.Unsubscribe{Topics: []string{subscribeTopic}})
		}
	}
}

func (c *ClientStruct) dispatch(subscribeTopic string, m Message) {
	c.subscriptionsMutex.RLock()
	handlers := slices.Collect(maps.Values(c.subscriptions[subscribeTopic]))
	c.subscriptionsMutex.RUnlock()

	for _, h := range handlers {
		h(m)
	}
}

func subscribeOptions(subscribeTopic string) paho.SubscribeOptions {
	return paho.SubscribeOptions{
		Topic: subscribeTopic,
		QoS:   byte(1),
	}
}

func logWrapper(clientName string, messageHandler MessageHandler) MessageHandler {
	return func(message Message) {
		// only log first 80 chars of payload
		const maxLen = 80
		pl := make([]byte, 0, maxLen+3)
		if len(message.Payload()) > maxLen {
			pl = append(pl, message.Payload()[:maxLen]...)
			pl = append(pl, "..."...)
		} else {
			pl = append(pl, message.Payload()...)
		}

		log.Printf("mqttClient[%s]: received: %s %s", clientName, message.Topic(), pl)
		messageHandler(message)
	}
}
