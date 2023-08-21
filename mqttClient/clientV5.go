package mqttClient

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"log"
	"net/url"
	"time"
)

func CreateV5(
	cfg Config,
) (client *ClientStruct) {
	ctx, cancel := context.WithCancel(context.Background())
	client = &ClientStruct{
		cfg:      cfg,
		shutdown: make(chan struct{}),

		router: paho.NewStandardRouter(),
		ctx:    ctx,
		cancel: cancel,
	}

	// configure mqtt library
	client.cliCfg = autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{cfg.Broker()},
		KeepAlive:         uint16(cfg.KeepAlive().Seconds()),
		ConnectRetryDelay: cfg.ConnectRetryDelay(),
		ConnectTimeout:    cfg.ConnectTimeout(),
		OnConnectionUp:    client.onConnectionUp(),
		OnConnectError: func(err error) {
			log.Printf("mqttClientV5[%s]: connection error: %s", cfg.Name(), err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.ClientId(),
			Router:   client.router,
		},
	}

	// setup logging
	if cfg.LogDebug() {
		prefix := fmt.Sprintf("mqttClientV5[%s]: ", cfg.Name())
		client.cliCfg.Debug = logger{prefix: prefix + "autoPaho: "}
		client.cliCfg.PahoDebug = logger{prefix: prefix + "paho: "}
	}

	// configure login
	if user := cfg.User(); len(user) > 0 {
		client.cliCfg.SetUsernamePassword(user, []byte(cfg.Password()))
	}

	// setup availability topic using will
	if client.AvailabilityEnabled() {
		client.cliCfg.SetWillMessage(client.GetAvailabilityTopic(), []byte(availabilityOffline), cfg.Qos(), true)
	}

	return
}

func (c *ClientStruct) Run() {
	// start connection manager
	var err error
	c.cm, err = autopaho.NewConnection(c.ctx, c.cliCfg)
	if err != nil {
		panic(err) // never happens
	}
}

func (c *ClientStruct) onConnectionUp() func(*autopaho.ConnectionManager, *paho.Connack) {
	return func(cm *autopaho.ConnectionManager, conack *paho.Connack) {
		log.Printf("mqttClientV5[%s]: connection is up", c.cfg.Name())
		// publish online
		if c.AvailabilityEnabled() {
			go func() {
				_, err := cm.Publish(c.ctx, c.availabilityMsg(availabilityOnline))
				if err != nil {
					log.Printf("mqttClientV5[%s]: error during publish: %s", c.cfg.Name(), err)
				}
			}()
		}
		// subscribe topics
		if len(c.subscriptions) > 0 {
			if _, err := cm.Subscribe(c.ctx, &paho.Subscribe{
				Subscriptions: func() (ret map[string]paho.SubscribeOptions) {
					c.subscriptionsMutex.RLock()
					defer c.subscriptionsMutex.RUnlock()
					ret = make(map[string]paho.SubscribeOptions, len(c.subscriptions))

					subOpts := paho.SubscribeOptions{QoS: c.cfg.Qos()}
					for _, s := range c.subscriptions {
						ret[s.subscribeTopic] = subOpts
					}
					return
				}(),
			}); err != nil {
				log.Printf("mqttClientV5[%s]: failed to subscribe: %s", c.cfg.Name(), err)
			}
		}

	}
}

func (c *ClientStruct) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if c.AvailabilityEnabled() {
		ctx, cancel := context.WithTimeout(c.ctx, time.Second)
		defer cancel()
		_, err := c.cm.Publish(ctx, c.availabilityMsg(availabilityOffline))
		if err != nil {
			log.Printf("mqttClientV5[%s]: error during publish: %s", c.cfg.Name(), err)
		}
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second)
	defer cancel()
	if err := c.cm.Disconnect(ctx); err != nil {
		log.Printf("mqttClientV5[%s]: error during disconnect: %s", c.cfg.Name(), err)
	}

	// cancel main context
	c.cancel()

	log.Printf("mqttClientV5[%s]: shutdown completed", c.cfg.Name())
}

func (c *ClientStruct) Publish(topic string, payload []byte, qos byte, retain bool) (err error) {
	_, err = c.cm.Publish(c.ctx, &paho.Publish{
		QoS:     qos,
		Topic:   ReplaceTemplate(topic, c.cfg),
		Payload: payload,
		Retain:  retain,
	})
	return
}

func (c *ClientStruct) availabilityMsg(payload string) *paho.Publish {
	return &paho.Publish{
		QoS:     c.cfg.Qos(),
		Topic:   c.GetAvailabilityTopic(),
		Payload: []byte(payload),
		Retain:  availabilityRetain,
	}
}
