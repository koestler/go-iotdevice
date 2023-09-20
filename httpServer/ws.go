package httpServer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/koestler/go-iotdevice/config"
	"log"
	"net"
	"time"
)

type outputMessage struct {
	Values map[string]map[string]valueResponse `json:"values"`
}

type authMessage struct {
	AuthToken string `json:"authToken"`
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

		// the follow line uses a loop variable; it must be outside the closure
		r.GET(relativePath, func(c *gin.Context) {
			authenticated := make(chan bool, 1)

			// no authentication needed for public views
			if view.IsPublic() {
				authenticated <- true
			}

			conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
			if err != nil {
				log.Printf("httpServer: %s%s: error during upgrade: %s", r.BasePath(), relativePath, err)
				if err := conn.Close(); err != nil {
					log.Printf("httpServer: %s%s: error during close: %s", r.BasePath(), relativePath, err)
				}
				return
			}
			log.Printf("httpServer: %s%s: connection established to %s", r.BasePath(), relativePath, c.ClientIP())

			subscriptionCtx, subscriptionCancel := context.WithCancel(context.Background())
			go func() {
				defer subscriptionCancel()
				defer conn.Close()
				defer close(authenticated)
				if env.Config.LogDebug() {
					defer log.Printf("httpServer: %s%s: connection closed to %s", r.BasePath(), relativePath, c.ClientIP())
				}

				authenticationCompleted := view.IsPublic()

				for {
					msg, op, err := wsutil.ReadClientData(conn)
					if op == ws.OpClose {
						return
					}
					if err != nil {
						return
					}
					if env.Config.LogDebug() {
						log.Printf("httpServer: %s%s: message received: %s", r.BasePath(), relativePath, msg)
					}
					if !authenticationCompleted {
						var authMsg authMessage
						if err := json.Unmarshal(msg, &authMsg); err == nil {
							if user, err := checkToken(authMsg.AuthToken, env.Authentication.JwtSecret()); err == nil {
								log.Printf("httpServer: %s%s: user=%s authenticated", r.BasePath(), relativePath, user)
								authenticated <- isViewAuthenticatedByUser(view, user, true)
								authenticationCompleted = true
							}
						}
					}
				}
			}()

			go wsValuesSender(
				env, view, conn, authenticated, subscriptionCtx,
				fmt.Sprintf("httpServer: %s%s", r.BasePath(), relativePath),
			)
		})
		if env.Config.LogConfig() {
			log.Printf("httpServer: GET %s%s -> setup websocket for view", r.BasePath(), relativePath)
		}
	}
}

func wsValuesSender(
	env *Environment,
	view config.ViewConfig,
	conn net.Conn,
	authenticated <-chan bool,
	ctx context.Context,
	logPrefix string,

) {
	filter := getFilter(view.Devices())
	subscription := env.StateStorage.Subscribe(ctx, filter)

	writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(writer)

	if env.Config.LogDebug() {
		defer log.Printf("%s: tx routine closed", logPrefix)
	}

	// wait for authentication, or until auth closes due to disconnect
	if !<-authenticated {
		// authentication failed, do not send anything
		return
	}

	// send all values after initial connect
	{
		values := env.StateStorage.GetStateFiltered(filter)
		response := compile2DResponse(values)
		if err := wsSendValuesResponse(writer, encoder, response); err != nil {
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
					if err := wsSendValuesResponse(writer, encoder, values); err != nil {
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

func wsSendValuesResponse(writer *wsutil.Writer, encoder *jsoniter.Encoder, values map[string]map[string]valueResponse) error {
	message := outputMessage{
		Values: values,
	}
	if err := encoder.Encode(&message); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
