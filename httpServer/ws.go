package httpServer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/mileusna/useragent"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const wsSendTimeout = 5 * time.Second
const wsSendInterval = 250 * time.Millisecond

// registers / values maps use deviceName as the first dimension and registerName as the second dimension.
type outputMessage struct {
	Registers map[string]map[string]registerResponse `json:"registers,omitempty"`
	Values    map[string]map[string]valueResponse    `json:"values,omitempty"`
}

type authMessage struct {
	AuthToken string `json:"authToken"`
}

// setupViewWs godoc
// @Summary Realtime values websocket
// @Description Websocket that sends all registers and values initially and sends updates of changed values subsequently.
// @Param viewName path string true "View name as provided by the config endpoint"
// @Produce json
// @success 200 {array} outputMessage
// @Failure 404 {object} ErrorResponse
// @Router /views/{viewName}/ws [get]
// @Security ApiKeyAuth
func setupValuesWs(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, view := range env.Views {
		relativePath := "views/" + view.Name() + "/ws"
		logPrefix := fmt.Sprintf("httpServer: %s%s", r.BasePath(), relativePath)
		viewFilter := getViewValueFilter(view.Devices())

		// the follow line uses a loop variable; it must be outside the closure
		r.GET(relativePath, func(c *gin.Context) {
			var websocketAcceptOptions = websocket.AcceptOptions{
				CompressionMode: websocket.CompressionContextTakeover,
			}

			ua := useragent.Parse(c.GetHeader("User-Agent"))
			if env.Config.LogDebug() {
				log.Printf("%s: User-Agent: %s", logPrefix, c.GetHeader("User-Agent"))
			}
			if ua.IsIOS() || ua.IsSafari() {
				// Safari is know to not work with fragmented compressed websockets
				// disable context takeover as a work around
				if env.Config.LogDebug() {
					log.Printf("%s: ios/safari detected, disable compression", logPrefix)
				}
				websocketAcceptOptions = websocket.AcceptOptions{
					CompressionMode: websocket.CompressionDisabled,
				}
			}

			conn, err := websocket.Accept(c.Writer, c.Request, &websocketAcceptOptions)
			defer func() {
				err := conn.CloseNow()
				if env.Config.LogDebug() {
					log.Printf("%s: error during close: %s", logPrefix, err)
				}
			}()
			if err != nil {
				log.Printf("%s: error during upgrade: %s", logPrefix, err)
				return
			} else if env.Config.LogDebug() {
				log.Printf("%s: connection established to %s", logPrefix, c.ClientIP())
			}

			senderCtx, senderCancel := context.WithCancel(c)
			defer senderCancel()

			startValueSenderOnce := sync.OnceFunc(func() {
				startValuesSender(senderCtx, env, viewFilter, conn, logPrefix)
			})

			if view.IsPublic() {
				// do not wait for auth message, start sender immediately
				startValueSenderOnce()
			}

			for {
				mt, msg, err := conn.Read(c)
				if err != nil {
					if env.Config.LogDebug() {
						log.Printf("%s: read error: %s", logPrefix, err)
					}
					return
				} else if env.Config.LogDebug() {
					log.Printf("%s: message received: mt=%d, msg=%s", logPrefix, mt, msg)
				}

				if mt == websocket.MessageText {
					var authMsg authMessage
					if err := json.Unmarshal(msg, &authMsg); err == nil {
						if user, err := checkToken(authMsg.AuthToken, env.Authentication.JwtSecret()); err == nil {
							log.Printf("httpServer: %s%s: user=%s authenticated", r.BasePath(), relativePath, user)
							if isViewAuthenticatedByUser(view, user, true) {
								startValueSenderOnce()
							}
						}
					}
				}

			}
		})
		if env.Config.LogConfig() {
			log.Printf("httpServer: GET %s%s -> setup websocket for view", r.BasePath(), relativePath)
		}
	}
}

func startValuesSender(
	ctx context.Context,
	env *Environment,
	viewFilter dataflow.ValueFilterFunc,
	conn *websocket.Conn,
	logPrefix string,

) {
	if env.Config.LogDebug() {
		log.Printf("%s: start value sender", logPrefix)
		defer log.Printf("%s: tx routine closed", logPrefix)
	}

	pv := &packedValues{
		registers: make(map[string]map[string]registerResponse),
		values:    make(map[string]map[string]valueResponse),
	}

	// subscribe to the storage and update the packet values
	go func() {
		subscription := env.StateStorage.SubscribeSendInitial(ctx, viewFilter)

		sentRegisters := make(map[string]map[string]registerResponse)

		// this loop must never block long. Otherwise, the stateStorage is stalled.
		for v := range subscription.Drain() {
			pv.mu.Lock()

			// only append registers, if it was not sent before
			if append2DRegisterResponse(sentRegisters, v) {
				append2DRegisterResponse(pv.registers, v)
			}
			append2DValueResponse(pv.values, v)
			pv.mu.Unlock()
		}
	}()

	// send the packet values to the websocket connection
	go func() {
		ticker := time.NewTicker(wsSendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msg, err := pv.encodeAndReset()
				if err != nil {
					log.Printf("%s: error while encoding values: %s", logPrefix, err)
					continue
				}
				if msg == nil {
					continue
				}

				err = writeMessage(ctx, conn, msg)
				if err != nil {
					log.Printf("%s: error while sending values: %s", logPrefix, err)
				}
			}
		}
	}()
}

func writeMessage(ctx context.Context, conn *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, wsSendTimeout)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, msg)
}

type packedValues struct {
	mu        sync.Mutex
	registers map[string]map[string]registerResponse
	values    map[string]map[string]valueResponse
}

func (pv *packedValues) encodeAndReset() (msg []byte, err error) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if len(pv.registers) == 0 && len(pv.values) == 0 {
		return nil, nil // nothing to send
	}

	defer clear(pv.registers)
	defer clear(pv.values)

	om := outputMessage{
		Registers: pv.registers,
		Values:    pv.values,
	}

	return json.Marshal(om)
}
