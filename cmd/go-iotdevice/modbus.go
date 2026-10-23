package main

import (
	"github.com/koestler/go-iotdevice/v3/config"
	"github.com/koestler/go-iotdevice/v3/modbus"
	"github.com/koestler/go-iotdevice/v3/pool"
	"log"
)

func runModbus(
	cfg *config.Config,
) (modbusPool *pool.Pool[*modbus.ModbusStruct]) {
	// run pool
	modbusPool = pool.RunPool[*modbus.ModbusStruct]()

	for _, mbCfg := range cfg.Modbus() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"modbus[%s]: start: device='%s', baudRate=%d, readTimeout=%s",
				mbCfg.Name(), mbCfg.Device(), mbCfg.BaudRate(), mbCfg.ReadTimeout(),
			)
		}
		if mb, err := modbus.New(mbCfg); err != nil {
			log.Printf("modbus[%s]: start failed: %s", mbCfg.Name(), err)
		} else {
			modbusPool.Add(mb)
		}
	}

	return
}
