package httpDevice

import "github.com/koestler/go-iotdevice/config"

type Implementation interface {
	GetPath() string
	HandleResponse(body []byte) error
	GetCategorySort(category string) int
}

func implementationFactory(ds *DeviceStruct) Implementation {
	switch k := ds.httpConfig.Kind(); k {
	case config.HttpTeracomKind:
		return &TeracomDevice{ds}
	case config.HttpShelly3mKind:
		return &TeracomDevice{}
	default:
		panic("unimplemented kind: " + k.String())
	}
}
