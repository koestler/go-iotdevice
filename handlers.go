package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vehttp"
	"io"
	"net/http"
	"strconv"
)

//go:generate frontend/install
//go:generate go-bindata -prefix "frontend/" -pkg main "frontend"

var HttpRoutes = vehttp.Routes{
	vehttp.Route{
		"Index",
		"GET",
		"/",
		HttpHandleAssetsGet,
	},
	vehttp.Route{
		"DeviceIndex",
		"GET",
		"/device/",
		HttpHandleDeviceIndex,
	},
	vehttp.Route{
		"DeviceIndex",
		"GET",
		"/device/{DeviceId:[0-9]+}",
		HttpHandleDeviceGet,
	},
	vehttp.Route{
		"Assets",
		"GET",
		"/assets/{Path}",
		HttpHandleAssetsGet,
	},
}

type jsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
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

func HttpHandleAssetsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	path := vars["Path"]
	if path == "" {
		path = "index.html"
	}

	if bs, err := Asset(path); err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 asset not found\n")
	} else {
		var reader = bytes.NewBuffer(bs)
		io.Copy(w, reader)
	}
}
