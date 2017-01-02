package vehttp

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/koestler/go-ve-sensor/vedata"
	"net/http"
	"strconv"
)

var HttpRoutes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"DeviceIndex",
		"GET",
		"/device/",
		HttpHandleDeviceIndex,
	},
	Route{
		"DeviceIndex",
		"GET",
		"/device/{DeviceId:[0-9]+}",
		HttpHandleDeviceGet,
	},
}

type jsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the go-ve-sensor server!\n")
}

func writeJsonHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
}

func jsonWriteNotFound(w http.ResponseWriter) {
	jsonWriteError(w, "Object Not Found")
}

func jsonWriteError(w http.ResponseWriter, text string) {
	writeJsonHeaders(w, http.StatusNotFound)

	ret := jsonErr{Code: http.StatusNotFound, Text: text}

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		panic(err)
	}
}

func HttpHandleDeviceIndex(w http.ResponseWriter, r *http.Request) {
	deviceIds := vedata.ReadDeviceIds()

	writeJsonHeaders(w, http.StatusOK)

	b, err := json.MarshalIndent(deviceIds, "", "    ")
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

func HttpHandleDeviceGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var deviceIdInt int
	var err error
	if deviceIdInt, err = strconv.Atoi(vars["DeviceId"]); err != nil {
		panic(err)
	}

	deviceId := vedata.DeviceId(deviceIdInt)
	device, err := deviceId.ReadDevice()

	if err == nil {
		writeJsonHeaders(w, http.StatusOK)
		b, err := json.MarshalIndent(device, "", "    ")
		if err != nil {
			panic(err)
		}
		w.Write(b)
	} else {
		jsonWriteNotFound(w)
	}
}
