// Package mqttClient implements a MQTT v5 client with automatic reconnects, does  and subscription management.
// It is a wrapper around the Eclipse Paho Go MQTT library which adds the following features:
// Both Subscribe and Publish are guaranteed to never block, even if the broker is not reachable.
// For Publishing messages, a FIFO queue with a limited size is used.
// Subscribing to topics is done automatically after reconnecting, and the library keeps track of the subscriptions,
// and allows to remove them again.
//
// Contract:
// - Publish / Subscribe never block (you can run them inside handlers).
// - Handlers are expected to never block (they are called inside the central run routine).
package mqttClient

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/internal/queue"
	"github.com/koestler/go-list"
)

const availabilityOnline = "online"
const availabilityOffline = "offline"

type ClientStruct struct {
	cfg Config

	cliCfg autopaho.ClientConfig

	subscriptionChannel chan subscriptionRequest
	dispatchChannel     chan Message
	connectedChannel    chan *autopaho.ConnectionManager // signals (re)connection, used to resubscribe
	shutdownChannel     chan chan struct{}

	nextSubscriptionID atomic.Uint64

	brokerOps  *list.LList[brokerOp] // ordered queue of subscribe / unsubscribe operations
	brokerWake chan struct{}         // signals the runBrokerOps routine that brokerOps is not empty

	publishBacklog *queue.FIFO[paho.Publish]        // queue of messages to be published with a maximum size
	publishWake    chan *autopaho.ConnectionManager // signals (re)connection to the runPublish routine

	ctx    context.Context
	cancel context.CancelFunc
}

type subscriptionRequest struct {
	id      uint64
	route   string
	handler MessageHandler // not nil -> subscribe, nil -> unsubscribe
}

type brokerOp struct {
	cm        *autopaho.ConnectionManager
	subscribe bool
	topics    []string
}

func NewV5(
	cfg Config,
) (client *ClientStruct) {
	ctx, cancel := context.WithCancel(context.Background())
	client = &ClientStruct{
		cfg: cfg,

		subscriptionChannel: make(chan subscriptionRequest),
		dispatchChannel:     make(chan Message),
		connectedChannel:    make(chan *autopaho.ConnectionManager, 1),
		shutdownChannel:     make(chan chan struct{}),

		brokerOps:  list.NewL[brokerOp](),
		brokerWake: make(chan struct{}, 1),

		publishBacklog: queue.NewFIFO[paho.Publish](cfg.MaxBacklogSize()),
		publishWake:    make(chan *autopaho.ConnectionManager, 1),

		ctx:    ctx,
		cancel: cancel,
	}

	// configure MQTT library
	client.cliCfg = autopaho.ClientConfig{
		ServerUrls:       []*url.URL{cfg.Broker()},
		KeepAlive:        uint16(cfg.KeepAlive().Seconds()),
		ReconnectBackoff: autopaho.NewConstantBackoff(cfg.ConnectRetryDelay()),
		ConnectTimeout:   cfg.ConnectTimeout(),
		OnConnectionUp:   client.onConnectionUp,
		OnConnectError: func(err error) {
			log.Printf("mqttClient[%s]: connection error: %s", cfg.Name(), err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID:          cfg.ClientId(),
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){client.onPublishReceived},
		},
	}

	// set up logging
	if cfg.LogDebug() {
		prefix := fmt.Sprintf("mqttClient[%s]: ", cfg.Name())
		client.cliCfg.Debug = logger{prefix: prefix + "debug: "}
		client.cliCfg.PahoDebug = logger{prefix: prefix + "paho: "}
	}

	// configure login
	if user := cfg.User(); len(user) > 0 {
		client.cliCfg.ConnectUsername = user
		client.cliCfg.ConnectPassword = []byte(cfg.Password())
	}

	// set up client availability topic using will
	if mCfg := cfg.AvailabilityClient(); mCfg.Enabled() {
		client.cliCfg.WillMessage = &paho.WillMessage{
			Retain:  mCfg.Retain(),
			QoS:     mCfg.Qos(),
			Topic:   cfg.AvailabilityClientTopic(),
			Payload: []byte(availabilityOffline),
		}
	}

	return
}

func (c *ClientStruct) Name() string {
	return c.cfg.Name()
}

func (c *ClientStruct) GetCtx() context.Context {
	return c.ctx
}

func (c *ClientStruct) Run() {
	go c.run()
	go c.runBrokerOps()
	go c.runPublish()

	// start connection manager
	_, _ = autopaho.NewConnection(c.ctx, c.cliCfg)
}

// Publish sends a message to the broker. It never blocks, even if the broker is not reachable.
func (c *ClientStruct) Publish(topic string, payload []byte, qos byte, retain bool) {
	if c.cfg.ReadOnly() {
		log.Printf("mqttClient[%s]: message dropped due to ReadOnly flag: topic=%s payload=%s", c.cfg.Name(), topic, payload)
		return
	}

	c.enqueuePublish(paho.Publish{
		QoS:     qos,
		Topic:   topic,
		Payload: payload,
		Retain:  retain,
	})
}

// AddRoute registers a handler for subscribeTopic and subscribes on the broker if needed.
// The function never blocks and returns immediately.
// The handler is called inside the central run routine and must be fast and never block (see MessageHandler).
// The returned function removes the handler again; the topic is unsubscribed once its last handler is removed.
func (c *ClientStruct) AddRoute(ctx context.Context, route string, handler MessageHandler) (remove func()) {
	if c.cfg.LogDebug() {
		log.Printf("mqttClient[%s]: adding route for topic: %s", c.cfg.Name(), route)
	}

	if c.cfg.LogMessages() {
		handler = logWrapper(c.cfg.Name(), handler)
	}

	id := c.nextSubscriptionID.Add(1)

	go func() {
		select {
		case <-ctx.Done(): // drop request in case the context is already cancelled
		case <-c.ctx.Done(): // drop request in case the client is already shutting down
		case c.subscriptionChannel <- subscriptionRequest{id: id, route: route, handler: handler}:
		}
	}()

	return func() {
		if c.cfg.LogDebug() {
			log.Printf("mqttClient[%s]: removing route for topic: %s", c.cfg.Name(), route)
		}
		go func() {
			select {
			case <-c.ctx.Done(): // drop request in case the client is already shutting down
			case c.subscriptionChannel <- subscriptionRequest{id: id, route: route}:
			}
		}()
	}
}

// Shutdown shuts down the client and disconnects from the broker.
// It blocks until the shutdown is complete.
func (c *ClientStruct) Shutdown() {
	done := make(chan struct{})
	select {
	case <-c.ctx.Done():
		return
	case c.shutdownChannel <- done:
	}
	<-done
}

func (c *ClientStruct) shutdown(cm *autopaho.ConnectionManager) {
	if cm != nil {
		// publish availability offline
		if c.cfg.AvailabilityClient().Enabled() {
			ctx, cancel := context.WithTimeout(c.ctx, time.Second)
			msg := c.availabilityMsg(availabilityOffline)
			_, err := cm.Publish(ctx, &msg)
			cancel()
			if err != nil {
				log.Printf("mqttClient[%s]: error publishing offline availability: %s", c.cfg.Name(), err)
			}
		}

		ctx, cancel := context.WithTimeout(c.ctx, time.Second)
		if err := cm.Disconnect(ctx); err != nil {
			log.Printf("mqttClient[%s]: error during disconnect: %s", c.cfg.Name(), err)
		}
		cancel()
	}

	// cancel main context
	c.cancel()

	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}
