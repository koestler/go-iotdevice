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

			// the follow line uses a loop variable; it must be outside the closure
			deviceFilter := dataflow.Filter{Devices: map[string]bool{deviceName: true}}
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}
				values := env.Storage.GetMap(deviceFilter)
				jsonGetResponse(c, compileResponse(values))
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> serve values as json", r.BasePath(), relativePath)
			}
		}
	}
}

func compileResponse(values dataflow.ValueMap) (response map[string]valueResponse) {
	response = make(map[string]valueResponse, len(values))
	for registerName, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			response[registerName] = numeric.Value()
		} else if text, ok := value.(dataflow.TextRegisterValue); ok {
			response[registerName] = text.Value()
		}
	}
	return
}

// setupValuesWs godoc
// @Summary Outputs the latest values of all the fields of a device.
// @ID valuesWs
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /values/{viewName}/{deviceName}.ws [get]
// @Security ApiKeyAuth
func setupValuesWs(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, deviceName := range view.DeviceNames() {

			device := env.DevicePoolInstance.GetDevice(deviceName)
			if device == nil {
				continue
			}

			relativePath := "values/" + view.Name() + "/" + deviceName + ".ws"

			// the follow line uses a loop variable; it must be outside the closure
			deviceFilter := dataflow.Filter{Devices: map[string]bool{deviceName: true}}
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c) {
					log.Printf("httpServer: %s%s: permission denied", r.BasePath(), relativePath)
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}

				conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
				if err != nil {
					log.Printf("httpServer: %s%s: error during upgrade: %s", r.BasePath(), relativePath, err)
					if err := conn.Close(); err != nil {
						log.Printf("httpServer: %s%s: error during close: %s", r.BasePath(), relativePath, err)
					}
					return
				}
				log.Printf("httpServer: %s%s: websocket connection established", r.BasePath(), relativePath)

				subscription := env.Storage.Subscribe(deviceFilter)

				go func() {
					defer subscription.Shutdown()
					defer conn.Close()

					for {
						msg, op, err := wsutil.ReadClientData(conn)
						if op == ws.OpClose {
							log.Printf("httpServer: %s%s: close connection", r.BasePath(), relativePath)
							return
						}
						if err != nil {
							log.Printf("httpServer: %s%s: error during read: %s", r.BasePath(), relativePath, err)
							return
						} else {
							log.Printf("httpServer: %s%s: message received: %s", r.BasePath(), relativePath, msg)
						}
					}
				}()

				go func() {
					writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
					encoder := json.NewEncoder(writer)

					// send all values after initial connect
					values := env.Storage.GetMap(deviceFilter)
					response := compileResponse(values)
					if err := wsSendResponse(writer, encoder, response); err != nil {
						log.Printf("httpServer: %s%s: error while sending value: %s", r.BasePath(), relativePath, err)
						return
					}

					// send updates
					for v := range subscription.GetOutput() {
						// compile response
						m := make(dataflow.ValueMap, 1)
						m[v.Register().Name()] = v
						response := compileResponse(m)

						// send response
						if err := wsSendResponse(writer, encoder, response); err != nil {
							log.Printf("httpServer: %s%s: error while sending value: %s", r.BasePath(), relativePath, err)
							return
						}
					}
					log.Printf("httpServer: %s%s: shutdown writing", r.BasePath(), relativePath)
				}()
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> setup websocket for values", r.BasePath(), relativePath)
			}
		}
	}
}

func wsSendResponse(writer *wsutil.Writer, encoder *json.Encoder, response map[string]valueResponse) error {
	log.Printf("websocket send: %v", response)
	if err := encoder.Encode(&response); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
