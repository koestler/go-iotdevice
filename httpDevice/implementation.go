package httpDevice

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"net/http"
)

type Implementation interface {
	GetPath() string
	HandleResponse(body []byte) error
	GetCategorySort(category string) int
	ControlValueRequest(value dataflow.Value) (*http.Request, error)
}

func implementationFactory(ds *DeviceStruct) Implementation {
	switch k := ds.httpConfig.Kind(); k {
	case config.HttpTeracomKind:
		return &TeracomDevice{ds}
	case config.HttpShellyEm3Kind:
		return &ShellyEm3Device{ds}
	default:
		panic("unimplemented kind: " + k.String())
	}
}
