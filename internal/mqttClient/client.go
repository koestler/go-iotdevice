// Package mqttClient implements a MQTT v5 client with automatic reconnects, subscription management, and routing.
// It is a wrapper around the Eclipse Paho Go MQTT library which adds the following features:
// Both AddRoute and Publish are guaranteed to never block, even if the broker is not reachable.
// For Publishing messages, a FIFO queue with a limited size is used.
// Subscribing to topics is done automatically after reconnecting, the library keeps track of the subscriptions
// and allows to remove them again.
//
// Contract:
// - Publish / AddRoute never block (you can run them inside handlers).
// - Handlers are expected to never block (they are called inside the central run routine).
//
// Behaviour worth knowing:
//   - All subscriptions use QoS 1.
//   - The context passed to AddRoute only cancels a request that is still pending; it does not remove a route
//     that is already registered. Use the returned remove function for that.
//   - Messages still in the publish backlog when Shutdown is called are discarded.
//   - Messages received while shutting down are dropped.
//   - If a message matches several routes (or several overlapping subscriptions), handlers may be called
//     once per matching route.
package mqttClient

import (
	"context"
	"fmt"
	"log"
	"math"
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

// time allowed to send the offline availability message and disconnect from the broker
const shutdownTimeout = 5 * time.Second

type ClientStruct struct {
	cfg Config

	cliCfg autopaho.ClientConfig

	// router loop channels
	routeChannel     chan routeRequest
	dispatchChannel  chan Message
	connectedChannel chan *autopaho.ConnectionManager // signals (re)connection, used to resubscribe

	nextSubscriptionID atomic.Uint64

	// subscribe loop channels
	subscribeBacklog *list.LList[brokerOp] // ordered queue of subscribe / unsubscribe operations
	subscribeWake    chan struct{}         // signals the runSubscribe routine that subscribeBacklog is not empty

	// publish loop channels
	publishBacklog *queue.FIFO[paho.Publish]        // queue of messages to be published with a maximum size
	publishWake    chan *autopaho.ConnectionManager // signals (re)connection to the runPublish routine

	// shutdown handling
	ctx       context.Context // cancelling this context starts the shutdown process
	cancelCtx context.CancelFunc
	done      chan struct{} // closed when the shutdown is complete
}

type routeRequest struct {
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
) (*ClientStruct, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &ClientStruct{
		cfg: cfg,

		routeChannel:     make(chan routeRequest),
		dispatchChannel:  make(chan Message),
		connectedChannel: make(chan *autopaho.ConnectionManager, 1),

		subscribeBacklog: list.NewL[brokerOp](),
		subscribeWake:    make(chan struct{}, 1),

		publishBacklog: queue.NewFIFO[paho.Publish](cfg.MaxBacklogSize()),
		publishWake:    make(chan *autopaho.ConnectionManager, 1),

		ctx:       ctx,
		cancelCtx: cancel,
		done:      make(chan struct{}),
	}

	// configure MQTT library
	client.cliCfg = autopaho.ClientConfig{
		ServerUrls:       []*url.URL{cfg.Broker()},
		KeepAlive:        keepAliveSeconds(cfg.KeepAlive()),
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
	if mCfg := cfg.AvailabilityClient(); mCfg.Enabled() && !cfg.ReadOnly() {
		client.cliCfg.WillMessage = &paho.WillMessage{
			Retain:  mCfg.Retain(),
			QoS:     mCfg.Qos(),
			Topic:   cfg.AvailabilityClientTopic(),
			Payload: []byte(availabilityOffline),
		}
	}

	// start connection manager using a background context, its context is cancelled by cm.Disconnect
	cm, err := autopaho.NewConnection(context.Background(), client.cliCfg)
	if err != nil {
		return nil, err
	}

	go client.runShutdown(cm)
	go client.runRouter(cm)
	go client.runSubscribe()
	go client.runPublish()

	return client, nil
}

// keepAliveSeconds converts d to the MQTT keep alive field, bounded to 1s..65535s.
func keepAliveSeconds(d time.Duration) uint16 {
	return uint16(min(max(d/time.Second, 1), math.MaxUint16))
}

func (c *ClientStruct) Name() string {
	return c.cfg.Name()
}

// GetCtx returns a context that is canceled when the client is shutting down.
func (c *ClientStruct) GetCtx() context.Context {
	return c.ctx
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
		case c.routeChannel <- routeRequest{id: id, route: route, handler: handler}:
		}
	}()

	return func() {
		if c.cfg.LogDebug() {
			log.Printf("mqttClient[%s]: removing route for topic: %s", c.cfg.Name(), route)
		}
		go func() {
			select {
			case <-c.ctx.Done(): // drop request in case the client is already shutting down
			case c.routeChannel <- routeRequest{id: id, route: route}:
			}
		}()
	}
}

// Shutdown shuts down the client and disconnects from the broker.
// If the client availability is enabled and the client is not read-only, it first publishes the offline message.
// It waits for that message and the disconnect to complete (at most shutdownTimeout), but does not wait for the backlog
// to be drained. It is safe to call multiple times and from within a message handler.
func (c *ClientStruct) Shutdown() {
	c.cancelCtx()
	<-c.done
}
