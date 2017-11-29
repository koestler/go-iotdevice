package httpServer

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/deviceDb"
)

func HandleDeviceIndex(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	devices := deviceDb.GetAll()

	// cache device index for 5 minutes
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJsonHeaders(w)

	b, err := json.MarshalIndent(devices, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)

	return nil;
}

func HandleDeviceGetRoundedValues(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	vars := mux.Vars(r)

	device, err := deviceDb.GetByName(vars["DeviceId"])
	if err != nil {
		return StatusError{404, err}
	}

	roundedValues := env.RoundedStorage.GetMap(dataflow.Filter{Devices: map[*deviceDb.Device]bool{device: true}})
	roundedValuesEssential := roundedValues.ConvertToEssential()

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(roundedValuesEssential, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)
	return nil
}
