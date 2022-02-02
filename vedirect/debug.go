package vedirect

import (
	"fmt"
	"log"
	"strings"
)

func (vd *Vedirect) debugPrintf(format string, v ...interface{}) {
	// check if debug output is enabled
	if !vd.logDebug {
		return
	}

	intro := strings.Split(format, "=")[0]

	if vd.logDebugIndent > 0 && strings.Contains(intro, " end") {
		vd.logDebugIndent -= 1
	}

	s := fmt.Sprintf(format, v...)
	s = strings.Replace(s, "\n", "\\n", -1)

	log.Print(strings.Repeat("  ", vd.logDebugIndent) + s)

	if vd.logDebugIndent < 64 && strings.Contains(intro, " begin") {
		vd.logDebugIndent += 1
	}
}
