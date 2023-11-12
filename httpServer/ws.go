package httpServer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/mileusna/useragent"
	"log"
	"nhooyr.io/websocket"
	"time"
)

// registers / values maps use deviceName as the first dimension and registerName as the second dimension.
type outputMessage struct {
	Operation string                                 `json:"op" example:"init"`
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
	for _, v := range env.Views {
		view := v
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
				err := conn.Close(websocket.StatusInternalError, "")
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

			valueSenderStarted := false
			startValueSenderOnce := func() {
				if valueSenderStarted {
					return
				}
				go wsValuesSender(env, viewFilter, conn, senderCtx, logPrefix)
				valueSenderStarted = true
			}

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

func wsValuesSender(
	env *Environment,
	viewFilter dataflow.ValueFilterFunc,
	conn *websocket.Conn,
	ctx context.Context,
	logPrefix string,

) {
	if env.Config.LogDebug() {
		log.Printf("%s: start value sender", logPrefix)
		defer log.Printf("%s: tx routine closed", logPrefix)
	}

	initial, subscription := env.StateStorage.SubscribeReturnInitial(ctx, viewFilter)

	// send all values after initial connect
	registers := compile2DRegisterResponse(initial)
	{
		values := compile2DValueResponse(initial)
		if err := wsSendResponse(ctx, conn, "init", registers, values); err != nil {
			log.Printf("%s: error while sending initial values: %s", logPrefix, err)
			return
		}
	}

	// send updates
	{
		// rate limit number of sent messages to 4 per second
		tickerDuration := 250 * time.Millisecond
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

		tickerRunning := true
		valuesC := subscription.Drain()

		newRegisters := make(map[string]map[string]registerResponse)
		newValues := make(map[string]map[string]valueResponse)
		for {
			select {
			case <-ticker.C:
				if len(newValues) > 0 {
					// there is data to send, send it
					if err := wsSendResponse(ctx, conn, "inc", newRegisters, newValues); err != nil {
						log.Printf("%s: error while sending value: %s", logPrefix, err)
						return
					}

					clear(newRegisters)
					clear(newValues)
				} else {
					// no data to send; stop timer
					ticker.Stop()
					tickerRunning = false
				}
			case v, ok := <-valuesC:
				if ok {
					if append2DRegisterResponse(registers, v) {
						append2DRegisterResponse(newRegisters, v)
					}
					append2DValueResponse(newValues, v)

					if !tickerRunning {
						ticker.Reset(tickerDuration)
						tickerRunning = true
					}
				} else {
					// subscription was shutdown, stop
					return
				}
			}
		}
	}
}

func wsSendResponse(
	ctx context.Context,
	conn *websocket.Conn,
	operation string,
	registers map[string]map[string]registerResponse,
	values map[string]map[string]valueResponse,
) error {
	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	err1 := json.NewEncoder(w).Encode(outputMessage{
		Operation: operation,
		Registers: registers,
		Values:    values,
	})
	err2 := w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
