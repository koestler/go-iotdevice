package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/modbus"
	"github.com/koestler/go-iotdevice/pool"
	"log"
)

func runModbus(
	cfg *config.Config,
) (modbusPoolInstance *pool.Pool[modbus.Modbus]) {
	// run pool
	modbusPoolInstance = pool.RunPool[modbus.Modbus]()

	for _, mbCfg := range cfg.Modbus() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"modbus[%s]: start: device='%s', baudRate=%d, readTimeout='%s'",
				mbCfg.Name(), mbCfg.Device(), mbCfg.BaudRate(), mbCfg.ReadTimeout(),
			)
		}
		if mb, err := modbus.Create(mbCfg); err != nil {
			log.Printf("modbus[%s]: start failed: %s", mbCfg.Name(), err)
		} else {
			modbusPoolInstance.Add(mb)
		}
	}

	return
}
