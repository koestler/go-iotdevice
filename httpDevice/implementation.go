package httpDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/types"
	"net/http"
)

type OnControlSuccess func()

type Implementation interface {
	GetPath() string
	HandleResponse(body []byte) error
	GetCategorySort(category string) int
	ControlValueRequest(value dataflow.Value) (*http.Request, OnControlSuccess, error)
}

func implementationFactory(ds *DeviceStruct) Implementation {
	switch k := ds.httpConfig.Kind(); k {
	case types.HttpTeracomKind:
		return &TeracomDevice{ds}
	case types.HttpShellyEm3Kind:
		return &ShellyEm3Device{ds}
	default:
		panic("unimplemented kind: " + k.String())
	}
}
