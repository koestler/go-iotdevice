package httpDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/types"
	"net/http"
)

type OnCommandSuccess func()

type Implementation interface {
	GetPath() string
	HandleResponse(body []byte) error
	GetCategorySort(category string) int
	CommandValueRequest(value dataflow.Value) (*http.Request, OnCommandSuccess, error)
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
