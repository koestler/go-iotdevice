package httpServer

import (
	"fmt"
	"log"
	"net/http"

	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/pkg/errors"
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
func setupValuesGetJson(mux *http.ServeMux, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, vd := range view.Devices() {
			pattern := "GET /api/v2/views/" + view.Name() + "/devices/" + vd.Name() + "/values"
			filter := getViewValueFilter([]ViewDeviceConfig{vd})
			handler := valuesGetJsonHandler(env, view, filter)

			mux.HandleFunc(pattern, authJwtMiddleware(env, gzipMiddleware(handler)))
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s -> serve values as json", pattern)
			}
		}
	}
}

func valuesGetJsonHandler(env *Environment, view ViewConfig, filter dataflow.ValueFilterFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check authorization
		if !isViewAuthenticated(view, r, true) {
			jsonErrorResponse(w, http.StatusForbidden, errors.New("User is not allowed here"))
			return
		}
		values := env.StateStorage.GetStateFiltered(filter)
		jsonGetResponse(w, r, compile1DValueResponse(values))
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
func setupValuesPatch(mux *http.ServeMux, env *Environment) {
	for _, view := range env.Views {
		for _, dn := range view.Devices() {
			pattern := "PATCH /api/v2/views/" + view.Name() + "/devices/" + dn.Name() + "/values"
			handler := valuesPatchHandler(env, view, dn.Name())

			mux.HandleFunc(pattern, authJwtMiddleware(env, handler))
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s -> setup value dispatcher", pattern)
			}
		}
	}
}

func valuesPatchHandler(env *Environment, view ViewConfig, deviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check authorization
		if !isViewAuthenticated(view, r, false) {
			jsonErrorResponse(w, http.StatusForbidden, errors.New("User is not allowed here"))
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErrorResponse(w, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
			return
		}

		// check all inputs
		inputs := make([]dataflow.Value, 0, len(req))
		for registerName, value := range req {
			register, ok := env.RegisterDbOfDevice(deviceName).GetByName(registerName)
			if !ok {
				jsonErrorResponse(w, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
				return
			}

			if !register.Writable() {
				jsonErrorResponse(w, http.StatusForbidden, fmt.Errorf("register %s is not writable", registerName))
				return
			}

			invalidType := func(t string) {
				jsonErrorResponse(w, http.StatusUnprocessableEntity, fmt.Errorf("expect type of %s to be a %s", registerName, t))
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

		w.WriteHeader(http.StatusOK)
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
