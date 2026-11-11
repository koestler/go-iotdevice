package mqttClient

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/url"
	"slices"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/v3/internal/queue"
)

const availabilityOnline = "online"
const availabilityOffline = "offline"

type ClientStruct struct {
	cfg Config

	cliCfg autopaho.ClientConfig
	router *paho.StandardRouter

	subscribeChannel   chan subscribeRequest
	unsubscribeChannel chan unsubscribeRequest
	dispatchChannel    chan dispatchRequest
	connectedChannel   chan *autopaho.ConnectionManager // signals (re)connection, used to resubscribe
	shutdownChannel    chan chan struct{}

	nextSubscriptionID atomic.Uint64

	publishBacklog *queue.FIFO[*paho.Publish]
	publishWake    chan *autopaho.ConnectionManager // signals (re)connection to the runPublish routine

	ctx    context.Context
	cancel context.CancelFunc
}

type subscribeRequest struct {
	id      uint64
	topic   string
	handler MessageHandler
}

type unsubscribeRequest struct {
	topic string
	id    uint64
}

type dispatchRequest struct {
	topic string
	msg   Message
}

func NewV5(
	cfg Config,
) (client *ClientStruct) {
	ctx, cancel := context.WithCancel(context.Background())
	client = &ClientStruct{
		cfg: cfg,

		router: paho.NewStandardRouter(),

		subscribeChannel:   make(chan subscribeRequest),
		unsubscribeChannel: make(chan unsubscribeRequest),
		dispatchChannel:    make(chan dispatchRequest),
		connectedChannel:   make(chan *autopaho.ConnectionManager, 1),
		shutdownChannel:    make(chan chan struct{}),

		publishBacklog: queue.NewFIFO[*paho.Publish](cfg.MaxBacklogSize()),
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
		OnConnectionUp:   client.onConnectionUp(),
		OnConnectError: func(err error) {
			log.Printf("mqttClient[%s]: connection error: %s", cfg.Name(), err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.ClientId(),
			Router:   client.router,
		},
	}

	// set up logging
	if cfg.LogDebug() {
		prefix := fmt.Sprintf("mqttClient[%s]: ", cfg.Name())
		client.cliCfg.Debug = logger{prefix: prefix + "debug: "}
		client.cliCfg.PahoDebug = logger{prefix: prefix + "paho: "}
		client.router.SetDebugLogger(logger{prefix: prefix + "router: "})
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
	go c.runPublish()

	// start connection manager
	_, _ = autopaho.NewConnection(c.ctx, c.cliCfg)
}

// run is the central routine; it is the only one touching the subscriptions.
func (c *ClientStruct) run() {
	name := c.cfg.Name()
	logDebug := c.cfg.LogDebug()

	var cm *autopaho.ConnectionManager
	subscriptions := make(map[string]map[uint64]MessageHandler) // topic -> routeId -> handler
	unsubscribed := make(map[uint64]struct{})                   // unsubscribe received before subscribe received

	for {
		select {
		case <-c.ctx.Done():
			c.publishBacklog.Close()
			return

		case cm = <-c.connectedChannel:
			topics := slices.Collect(maps.Keys(subscriptions))
			c.subscribeBroker(cm, topics...)

		case req := <-c.subscribeChannel:
			if _, ok := unsubscribed[req.id]; ok {
				// unsubscribe was received before subscribe, ignore this subscribe
				delete(unsubscribed, req.id)
				continue
			}

			handlers, exists := subscriptions[req.topic]
			if !exists {
				handlers = make(map[uint64]MessageHandler)
				subscriptions[req.topic] = handlers

				// the router has one dispatcher per topic; it forwards via the dispatch channel
				topic := req.topic
				if logDebug {
					log.Printf("mqttClient[%s]: registering router for topic: %s", name, topic)
				}
				c.router.RegisterHandler(topic, func(p *paho.Publish) {
					select {
					case <-c.ctx.Done():
					case c.dispatchChannel <- dispatchRequest{topic: topic, msg: Message{topic: p.Topic, payload: p.Payload}}:
					}
				})
				if logDebug {
					log.Printf("mqttClient[%s]: subscribing to topic: %s", name, topic)
				}
				c.subscribeBroker(cm, topic)
			}

			handlers[req.id] = req.handler

		case req := <-c.unsubscribeChannel:
			handlers, ok := subscriptions[req.topic]
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
			delete(subscriptions, req.topic)
			c.router.UnregisterHandler(req.topic)
			if cm != nil {
				if logDebug {
					log.Printf("mqttClient[%s]: unsubscribing from topic: %s", name, req.topic)
				}
				go func(topic string) {
					_, _ = cm.Unsubscribe(c.ctx, &paho.Unsubscribe{Topics: []string{topic}})
				}(req.topic)
			}

		case req := <-c.dispatchChannel:
			for _, h := range subscriptions[req.topic] {
				h(req.msg)
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

func (c *ClientStruct) subscribeBroker(cm *autopaho.ConnectionManager, topics ...string) {
	if cm == nil || len(topics) == 0 {
		return
	}
	opts := make([]paho.SubscribeOptions, 0, len(topics))
	for _, t := range topics {
		opts = append(opts,
			paho.SubscribeOptions{
				Topic: t,
				QoS:   byte(1),
			},
		)
	}

	// broker calls block until acknowledged, which needs the router to keep delivering; never do them inline
	go func() {
		if _, err := cm.Subscribe(c.ctx, &paho.Subscribe{Subscriptions: opts}); err != nil {
			log.Printf("mqttClient[%s]: failed to subscribe: %s", c.cfg.Name(), err)
		}
	}()
}

// AddRoute registers a handler for subscribeTopic and subscribes on the broker if needed.
// The function never blocks and returns immediately.
// The handler is called inside the central run routine and must be fast and never block (see MessageHandler).
// The returned function removes the handler again; the topic is unsubscribed once its last handler is removed.
func (c *ClientStruct) AddRoute(ctx context.Context, subscribeTopic string, handler MessageHandler) (remove func()) {
	if c.cfg.LogDebug() {
		log.Printf("mqttClient[%s]: adding route for topic: %s", c.cfg.Name(), subscribeTopic)
	}

	if c.cfg.LogMessages() {
		handler = logWrapper(c.cfg.Name(), handler)
	}

	id := c.nextSubscriptionID.Add(1)

	go func() {
		select {
		case <-ctx.Done(): // drop request in case the context is already cancelled
		case <-c.ctx.Done(): // drop request in case the client is already shutting down
		case c.subscribeChannel <- subscribeRequest{id: id, topic: subscribeTopic, handler: handler}:
		}
	}()

	return func() {
		if c.cfg.LogDebug() {
			log.Printf("mqttClient[%s]: removing route for topic: %s", c.cfg.Name(), subscribeTopic)
		}
		go func() {
			select {
			case <-c.ctx.Done(): // drop request in case the client is already shutting down
			case c.unsubscribeChannel <- unsubscribeRequest{topic: subscribeTopic, id: id}:
			}
		}()
	}
}

func (c *ClientStruct) onConnectionUp() func(*autopaho.ConnectionManager, *paho.Connack) {
	return func(cm *autopaho.ConnectionManager, connack *paho.Connack) {
		log.Printf("mqttClient[%s]: connection is up", c.cfg.Name())

		// hand over to the run routine which (re)subscribes all topics
		select {
		case <-c.ctx.Done():
		case c.connectedChannel <- cm:
		}

		if c.cfg.AvailabilityClient().Enabled() {
			c.enqueue(c.availabilityMsg(availabilityOnline))
		}

		// wake the publish routine; a pending wake-up is enough, so never block
		select {
		case c.publishWake <- cm:
		default:
		}
	}
}

// runPublish lives as long as the client; after every (re)connection it drains the backlog.
// A wake-up arriving while a previous drain is still failing stays buffered and is handled right after,
// so a reconnect can never be missed.
func (c *ClientStruct) runPublish() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case cm := <-c.publishWake:
			c.drainBacklog(cm)
		}
	}
}

// drainBacklog forwards the backlog to the broker until the backlog is closed or a publish fails.
func (c *ClientStruct) drainBacklog(cm *autopaho.ConnectionManager) {
	for e := range c.publishBacklog.Dequeue() {
		pr, err := cm.Publish(c.ctx, e.Value)

		// a response with an error reason code means the broker received but rejected the message;
		// retrying would block the backlog forever, so discard it
		if pr != nil && pr.ReasonCode >= 0x80 {
			log.Printf("mqttClient[%s]: message rejected by broker, discarding: topic=%s reason=%d: %v",
				c.cfg.Name(), e.Value.Topic, pr.ReasonCode, err,
			)
			e.Next(true)
			continue
		}

		if err != nil {
			// connection problem: keep the message at the front of the backlog and retry after reconnect
			e.Next(false)
			if c.cfg.LogDebug() {
				log.Printf("mqttClient[%s]: error during publish, keeping in backlog: %s", c.cfg.Name(), err)
			}
			return
		}
		e.Next(true)
	}
}

func (c *ClientStruct) enqueue(p *paho.Publish) {
	if err := c.publishBacklog.Enqueue(p); err != nil {
		log.Printf("mqttClient[%s]: message dropped: topic=%s: %s", c.cfg.Name(), p.Topic, err)
	}
}

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
			_, err := cm.Publish(ctx, c.availabilityMsg(availabilityOffline))
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

func (c *ClientStruct) Publish(topic string, payload []byte, qos byte, retain bool) {
	if c.cfg.ReadOnly() {
		log.Printf("mqttClient[%s]: message dropped due to ReadOnly flag: topic=%s payload=%s", c.cfg.Name(), topic, payload)
		return
	}

	c.enqueue(&paho.Publish{
		QoS:     qos,
		Topic:   topic,
		Payload: payload,
		Retain:  retain,
	})
}

func (c *ClientStruct) availabilityMsg(payload string) *paho.Publish {
	mCfg := c.cfg.AvailabilityClient()

	return &paho.Publish{
		QoS:     mCfg.Qos(),
		Topic:   c.cfg.AvailabilityClientTopic(),
		Payload: []byte(payload),
		Retain:  mCfg.Retain(),
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
