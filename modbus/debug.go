package modbus

import (
	"fmt"
	"log"
	"strings"
)

func (md *ModbusStruct) debugPrintf(format string, v ...interface{}) {
	// check if debug output is enabled
	if !md.cfg.LogDebug() {
		return
	}

	intro := strings.Split(format, "=")[0]

	if md.logDebugIndent > 0 && strings.Contains(intro, " end") {
		md.logDebugIndent -= 1
	}

	s := fmt.Sprintf(format, v...)
	s = strings.Replace(s, "\n", "\\n", -1)

	log.Printf("modbus[%s]: %s", md.cfg.Name(), strings.Repeat("  ", md.logDebugIndent)+s)

	if md.logDebugIndent < 64 && strings.Contains(intro, " begin") {
		md.logDebugIndent += 1
	}
}
