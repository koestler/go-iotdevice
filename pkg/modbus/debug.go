package modbus

import (
	"fmt"
	"log"
)

func (md *ModbusStruct) debugPrintf(format string, v ...interface{}) {
	// check if debug output is enabled
	if !md.cfg.LogDebug() {
		return
	}

	s := fmt.Sprintf(format, v...)
	log.Printf("modbus[%s]: %s", md.cfg.Name(), s)
}
