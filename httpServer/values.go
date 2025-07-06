package httpServer

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type valueResponse interface{}
type values1DResponse map[string]valueResponse

// setupValuesGetJson godoc
// @Summary List values
// @Description Outputs the latest values of all the registers of a device.
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {object} values1DResponse
// @Failure 404 {object} ErrorResponse
// @Router /views/{viewName}/devices/{deviceName}/values [get]
// @Security ApiKeyAuth
func setupValuesGetJson(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, vd := range view.Devices() {
			viewDevice := vd

			relativePath := "views/" + view.Name() + "/devices/" + viewDevice.Name() + "/values"

			filter := getViewValueFilter([]ViewDeviceConfig{viewDevice})
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c, true) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}
				values := env.StateStorage.GetStateFiltered(filter)
				jsonGetResponse(c, compile1DValueResponse(values))
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: GET %s%s -> serve values as json", r.BasePath(), relativePath)
			}
		}
	}
}

// setupValuesPatch godoc
// @Summary Set value
// @Description Sets a writable register to a certain value.
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200
// @Failure 404 {object} ErrorResponse
// @Router /views/{viewName}/devices/{deviceName}/values [patch]
// @Security ApiKeyAuth
func setupValuesPatch(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, dn := range view.Devices() {
			// the following line uses a loop variable; it must be outside the closure
			deviceName := dn.Name()

			relativePath := "views/" + view.Name() + "/devices/" + deviceName + "/values"
			r.PATCH(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c, false) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}

				var req map[string]interface{}
				if err := c.ShouldBindJSON(&req); err != nil {
					jsonErrorResponse(c, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
					return
				}

				// check all inputs
				inputs := make([]dataflow.Value, 0, len(req))
				for registerName, value := range req {
					register, ok := env.RegisterDbOfDevice(deviceName).GetByName(registerName)
					if !ok {
						jsonErrorResponse(c, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
						return
					}

					invalidType := func(t string) {
						jsonErrorResponse(c, http.StatusUnprocessableEntity, fmt.Errorf("expect type of %s to be a %s", registerName, t))
					}

					switch register.RegisterType() {
					case dataflow.TextRegister:
						if v, ok := value.(string); ok {
							inputs = append(inputs, dataflow.NewTextRegisterValue(deviceName, register, v))
						} else {
							invalidType("string")
							return
						}
					case dataflow.NumberRegister:
						if v, ok := value.(float64); ok {
							inputs = append(inputs, dataflow.NewNumericRegisterValue(deviceName, register, v))
						} else {
							invalidType("float")
							return
						}

					case dataflow.EnumRegister:
						if v, ok := value.(float64); ok {
							inputs = append(inputs, dataflow.NewEnumRegisterValue(deviceName, register, int(v)))
						} else {
							invalidType("float")
							return
						}
					}
				}

				// all ok, send inputs to storage
				for _, inp := range inputs {
					env.CommandStorage.Fill(inp)
				}
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: PATCH %s%s -> setup value dispatcher", r.BasePath(), relativePath)
			}
		}
	}
}

func compile1DValueResponse(values []dataflow.Value) (response values1DResponse) {
	response = make(map[string]valueResponse, len(values))
	for _, value := range values {
		response[value.Register().Name()] = value.GenericValue()
	}
	return
}

func append2DValueResponse(response map[string]map[string]valueResponse, value dataflow.Value) {
	d0 := value.DeviceName()
	d1 := value.Register().Name()

	if _, ok := response[d0]; !ok {
		response[d0] = make(map[string]valueResponse)
	}

	response[d0][d1] = value.GenericValue()
}

func getViewValueFilter(viewDevices []ViewDeviceConfig) dataflow.ValueFilterFunc {
	filters := make(map[string]dataflow.ValueFilterFunc)
	for _, vd := range viewDevices {
		filters[vd.Name()] = dataflow.RegisterValueFilter(vd.Filter())
	}

	return func(value dataflow.Value) bool {
		deviceName := value.DeviceName()
		if f, ok := filters[deviceName]; !ok {
			return false // device not included
		} else {
			return f(value)
		}
	}
}
