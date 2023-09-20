package httpServer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/config"
	"log"
	"nhooyr.io/websocket"
	"time"
)

type outputMessage struct {
	Values map[string]map[string]valueResponse `json:"values"`
}

type authMessage struct {
	AuthToken string `json:"authToken"`
}

var websocketAcceptOptions = websocket.AcceptOptions{
	CompressionMode: websocket.CompressionContextTakeover,
}

// setupViewWs godoc
// @Summary Websocket that sends all values initially and sends updates of changed values subsequently.
// @ID viewWs
// @Param viewName path string true "View name as provided by the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /views/{viewName}/ws [get]
// @Security ApiKeyAuth
func setupValuesWs(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		relativePath := "views/" + view.Name() + "/ws"
		logPrefix := fmt.Sprintf("httpServer: %s%s", r.BasePath(), relativePath)

		// the follow line uses a loop variable; it must be outside the closure
		r.GET(relativePath, func(c *gin.Context) {
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
				go wsValuesSender(env, view, conn, senderCtx, logPrefix)
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
	view config.ViewConfig,
	conn *websocket.Conn,
	ctx context.Context,
	logPrefix string,

) {
	if env.Config.LogDebug() {
		log.Printf("%s: start value sender", logPrefix)
		defer log.Printf("%s: tx routine closed", logPrefix)
	}

	filter := getFilter(view.Devices())
	subscription := env.StateStorage.Subscribe(ctx, filter)

	// send all values after initial connect
	{
		values := env.StateStorage.GetStateFiltered(filter)
		response := compile2DResponse(values)
		if err := wsSendValuesResponse(ctx, conn, response); err != nil {
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
		values := make(map[string]map[string]valueResponse, 1)
		for {
			select {
			case <-ticker.C:
				if len(values) > 0 {
					// there is data to send, send it
					if err := wsSendValuesResponse(ctx, conn, values); err != nil {
						log.Printf("%s: error while sending value: %s", logPrefix, err)
						return
					}
					values = make(map[string]map[string]valueResponse, 1)
				} else {
					// no data to send; stop timer
					ticker.Stop()
					tickerRunning = false
				}
			case v, ok := <-valuesC:
				if ok {
					append2DResponseValue(values, v)
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

func wsSendValuesResponse(ctx context.Context, conn *websocket.Conn, values map[string]map[string]valueResponse) error {
	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	err1 := json.NewEncoder(w).Encode(outputMessage{
		Values: values,
	})
	err2 := w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
