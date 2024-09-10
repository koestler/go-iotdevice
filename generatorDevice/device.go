package generatorDevice

import (
	"context"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
)

type Config interface {
}

type DeviceStruct struct {
	device.State

	stateStorage   *dataflow.ValueStorage
	commandStorage *dataflow.ValueStorage
}

func NewDevice(
	deviceConfig device.Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		stateStorage:   stateStorage,
		commandStorage: commandStorage,
	}

}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	return nil, false
}

func (c *DeviceStruct) Model() string {
	return "Generator Controller"
}

// func temperaturAndFireCheck(c Configuration, i Inputs) bool {
// 	return !i.FireDetected && i.IOAvailable &&
// 		i.EngineTemp >= c.EngineTempMin && i.EngineTemp <= c.EngineTempMax &&
// 		i.AirIntakeTemp >= c.AirIntakeTempMin && i.AirIntakeTemp <= c.AirIntakeTempMax &&
// 		i.AirExhaustTemp >= c.AirExhaustTempMin && i.AirExhaustTemp <= c.AirExhaustTempMax
// }

// func generatorOutputCheck(c Configuration, i Inputs) bool {
// 	return i.MessurementAvailable &&
// 		i.F >= c.FMin && i.F <= c.FMax &&
// 		i.U0 >= c.UMin && i.U0 <= c.UMax &&
// 		i.U1 >= c.UMin && i.U1 <= c.UMax &&
// 		i.U2 >= c.UMin && i.U2 <= c.UMax &&
// 		i.L0 <= c.PMax &&
// 		i.L1 <= c.PMax &&
// 		i.L2 <= c.PMax &&
// 		i.L0+i.L1+i.L2 <= c.PTotMax
// }
