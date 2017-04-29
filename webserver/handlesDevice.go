package webserver

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/koestler/go-ve-sensor/dataflow"
)

func HandleDeviceIndex(env *Environment, w http.ResponseWriter, r *http.Request) error {
	devices := dataflow.DevicesGet()
	/*
	deviceNames := make([]string, len(devices), len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}
	*/

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(devices, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)

	return nil;
}

func HandleDeviceGet(env *Environment, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	device, err := dataflow.DevicesGetByName(vars["DeviceId"])
	if err != nil {
		return StatusError{404, err}
	}

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(device, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)
	return nil
}
