package mqttClient

import (
	"context"
	"log"
	"sync"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/internal/queue"
)

type ClientStruct struct {
	cfg Config

	subscriptionsMutex sync.RWMutex
	subscriptions      []subscription

	cliCfg         autopaho.ClientConfig
	cm             *autopaho.ConnectionManager
	router         *paho.StandardRouter
	publishBacklog *queue.Fifo[*paho.Publish]

	ctx    context.Context
	cancel context.CancelFunc
}

type subscription struct {
	subscribeTopic string
	messageHandler MessageHandler
}

func (c *ClientStruct) Name() string {
	return c.cfg.Name()
}

func (c *ClientStruct) GetCtx() context.Context {
	return c.ctx
}

func (c *ClientStruct) AddRoute(subscribeTopic string, messageHandler MessageHandler) {
	s := subscription{subscribeTopic: subscribeTopic, messageHandler: messageHandler}

	if c.cfg.LogMessages() {
		s.messageHandler = func(message Message) {
			// only log first 80 chars of payload
			const maxLen = 80
			pl := make([]byte, 0, maxLen+3)
			if len(message.Payload()) > maxLen {
				pl = append(pl, message.Payload()[:maxLen]...)
				pl = append(pl, "..."...)
			} else {
				pl = append(pl, message.Payload()...)
			}

			log.Printf("mqttClient[%s]: received: %s %s", c.cfg.Name(), message.Topic(), pl)
			messageHandler(message)
		}
	}

	c.subscriptionsMutex.Lock()
	defer c.subscriptionsMutex.Unlock()
	c.subscriptions = append(c.subscriptions, s)

	// add route
	c.router.RegisterHandler(s.subscribeTopic, func(p *paho.Publish) {
		s.messageHandler(Message{
			topic:   p.Topic,
			payload: p.Payload,
		})
	})

	// send subscribe
	_, _ = c.cm.Subscribe(c.ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{s.pahoOptions()},
	})
}

func (s subscription) pahoOptions() paho.SubscribeOptions {
	return paho.SubscribeOptions{
		Topic: s.subscribeTopic,
		QoS:   byte(1),
	}
}
