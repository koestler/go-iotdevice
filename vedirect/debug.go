package vedirect

import (
	"log"
	"strings"
)

var indent = 0

func debugPrintf(format string, v ...interface{}) {
	if strings.Contains(format, "end") {
		indent -= 1
	}

	log.Printf(strings.Repeat("  ", indent)+format, v...)

	if strings.Contains(format, "begin") {
		indent += 1
	}
}
