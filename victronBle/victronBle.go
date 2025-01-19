package victronBle

import (
	"github.com/koestler/go-victron/vebleapi"
	"github.com/koestler/go-victron/veconst"
)

type VictronBle struct {
	adapter *vebleapi.Adapter
}

func New() (*VictronBle, error) {
	adapter, err := vebleapi.NewDefaultAdapter(veconst.BleManufacturerId, nil)
	if err != nil {
		return nil, err
	}

	return &VictronBle{
		adapter: adapter,
	}, nil
}
