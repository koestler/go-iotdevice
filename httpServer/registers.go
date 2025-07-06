package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

type registerResponse struct {
	Category    string         `json:"category" example:"Monitor"`
	Name        string         `json:"name" example:"PanelPower"`
	Description string         `json:"description" example:"Panel power"`
	Type        string         `json:"type" example:"numeric"`
	Enum        map[int]string `json:"enum,omitempty"`
	Unit        string         `json:"unit,omitempty" example:"W"`
	Sort        int            `json:"sort" example:"100"`
	Writable    bool           `json:"commandable" example:"false"` // json is kept at commandable for compatibility reasons
	// consider changing when going to majer version 4
}

// setupRegisters godoc
// @Summary List registers
// @Description Outputs information about all the available registers (not the current values) of the given device.
// @Description Depending on the device model (bmv, bluesolar) a different set of registers are available.
// @Description On some devices (e.g. TCW241), the registers are even dynamic and configurable on the device.
// @Description This endpoint outputs a list of fields including a name, a unit and a datatype.
// @ID registers
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {array} registerResponse
// @Failure 404 {object} ErrorResponse
// @Router /views/{viewName}/devices/{deviceName}/registers [get]
// @Security ApiKeyAuth
func setupRegisters(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, vd := range view.Devices() {
			viewDevice := vd

			deviceName := viewDevice.Name()

			relativePath := "views/" + view.Name() + "/devices/" + viewDevice.Name() + "/registers"
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c, true) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}

				registers := env.RegisterDbOfDevice(deviceName).GetAll()
				registers = dataflow.FilterRegisters(registers, viewDevice.Filter())
				dataflow.SortRegisterStructs(registers)

				setCacheControlPublic(c, 10*time.Second)
				jsonGetResponse(c, compile1DRegisterResponse(registers))
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: GET %s%s -> serve registers", r.BasePath(), relativePath)
			}
		}
	}
}

func createRegisterResponse(r dataflow.Register) registerResponse {
	return registerResponse{
		Category:    r.Category(),
		Name:        r.Name(),
		Description: r.Description(),
		Type:        r.RegisterType().String(),
		Enum:        r.Enum(),
		Unit:        r.Unit(),
		Sort:        r.Sort(),
		Writable:    r.Writable(),
	}
}

func compile1DRegisterResponse(registers []dataflow.RegisterStruct) (response []registerResponse) {
	response = make([]registerResponse, len(registers))
	for i, v := range registers {
		response[i] = createRegisterResponse(v)
	}
	return
}

func append2DRegisterResponse(m map[string]map[string]registerResponse, value dataflow.Value) (created bool) {
	d0 := value.DeviceName()
	d1 := value.Register().Name()

	if _, ok := m[d0]; !ok {
		m[d0] = make(map[string]registerResponse)
	}

	if _, ok := m[d0][d1]; !ok {
		m[d0][d1] = createRegisterResponse(value.Register())
		return true
	}

	return false
}
