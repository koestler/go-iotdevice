package httpServer

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

type valueResponse interface{}

// setupValuesJson godoc
// @Summary Outputs the latest values of all the fields of a device.
// @ID valuesJson
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /values/{viewName}/{deviceName}.json [get]
// @Security ApiKeyAuth
func setupValuesJson(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, deviceName := range view.DeviceNames() {

			device := env.DevicePoolInstance.GetDevice(deviceName)
			if device == nil {
				continue
			}

			relativePath := "values/" + view.Name() + "/" + deviceName + ".json"

			// the following line uses a loop variable; it must be outside the closure
			deviceFilter := dataflow.Filter{Devices: map[string]bool{deviceName: true}}
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}
				values := env.Storage.GetSlice(deviceFilter)
				jsonGetResponse(c, compile1DResponse(values))
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> serve values as json", r.BasePath(), relativePath)
			}
		}
	}
}

func compile1DResponse(values []dataflow.Value) (response map[string]valueResponse) {
	response = make(map[string]valueResponse, len(values))
	for _, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			response[value.Register().Name()] = numeric.Value()
		} else if text, ok := value.(dataflow.TextRegisterValue); ok {
			response[value.Register().Name()] = text.Value()
		}
	}
	return
}

type authMessage struct {
	AuthToken string `json:"authToken"`
}

// setupValuesWs godoc
// @Summary Websocket that sends all values initially and sends updates of changed values subsequently.
// @ID valuesWs
// @Param viewName path string true "View name as provided by the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /values/{viewName}/ws [get]
// @Security ApiKeyAuth
func setupValuesWs(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		relativePath := "values/" + view.Name() + "/ws"
		deviceFilter := getDeviceFilter(view.DeviceNames())

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

			subscription := env.Storage.Subscribe(deviceFilter)

			go func() {
				defer subscription.Shutdown()
				defer conn.Close()
				defer close(authenticated)
				defer log.Printf("httpServer: %s%s: connection closed to %s", r.BasePath(), relativePath, c.ClientIP())

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
							if user, err := checkToken(authMsg.AuthToken, env.Auth.JwtSecret()); err == nil {
								log.Printf("httpServer: %s%s: user=%s authenticated", r.BasePath(), relativePath, user)
								authenticated <- isViewAuthenticatedByUser(view, user)
								authenticationCompleted = true
							}
						}
					}
				}
			}()

			go func() {
				writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
				encoder := json.NewEncoder(writer)

				// wait for authentication
				if !<-authenticated {
					// authentication failed, do not send anything
					return
				}

				// send all values after initial connect
				{
					values := env.Storage.GetSlice(deviceFilter)
					response := compile2DResponse(values)
					if err := wsSendResponse(writer, encoder, response); err != nil {
						log.Printf("httpServer: %s%s: error while sending initial values: %s", r.BasePath(), relativePath, err)
						return
					}
				}

				// send updates
				{
					// rate limit number of sent messages to 4 per second
					tickerDuration := 250 * time.Millisecond
					ticker := time.NewTicker(tickerDuration)
					tickerRunning := true
					valuesC := subscription.GetOutput()
					response := make(map[string]map[string]valueResponse, 1)
					for {
						select {
						case <-ticker.C:
							if len(response) > 0 {
								// there is data to send, send it
								if err := wsSendResponse(writer, encoder, response); err != nil {
									log.Printf("httpServer: %s%s: error while sending value: %s", r.BasePath(), relativePath, err)
									return
								}
								response = make(map[string]map[string]valueResponse, 1)
							} else {
								// no data to send; stop timer
								ticker.Stop()
								tickerRunning = false
							}
						case v, ok := <-valuesC:
							if ok {
								append2DResponseValue(response, v)
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
			}()
		})
		if env.Config.LogConfig() {
			log.Printf("httpServer: %s%s -> setup websocket for values", r.BasePath(), relativePath)
		}
	}
}

func compile2DResponse(values []dataflow.Value) (response map[string]map[string]valueResponse) {
	response = make(map[string]map[string]valueResponse, 1)
	for _, value := range values {
		append2DResponseValue(response, value)
	}
	return
}

func append2DResponseValue(response map[string]map[string]valueResponse, value dataflow.Value) {
	d0 := value.DeviceName()
	d1 := value.Register().Name()

	if _, ok := response[d0]; !ok {
		response[d0] = make(map[string]valueResponse, 1)
	}

	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		response[d0][d1] = numeric.Value()
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		response[d0][d1] = text.Value()
	}
}

func getDeviceFilter(deviceNames []string) dataflow.Filter {
	devices := make(map[string]bool, len(deviceNames))
	for _, deviceName := range deviceNames {
		devices[deviceName] = true
	}
	return dataflow.Filter{Devices: devices}
}

func wsSendResponse(writer *wsutil.Writer, encoder *json.Encoder, response map[string]map[string]valueResponse) error {
	if err := encoder.Encode(&response); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
