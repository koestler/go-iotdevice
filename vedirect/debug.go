package vedirect

import (
	"log"
	"strings"
	"fmt"
	"github.com/koestler/go-ve-sensor/config"
)

var indent = 0

func debugPrintf(format string, v ...interface{}) {
	// check if debug output is enabled
	if !config.VedirectConfig.DebugPrint {
		return
	}

	intro := strings.Split(format, "=")[0]

	if indent > 0 && strings.Contains(intro, " end") {
		indent -= 1
	}

	s := fmt.Sprintf(format, v...)
	s = strings.Replace(s, "\n", "\\n", -1)

	log.Print(strings.Repeat("  ", indent) + s)

	if indent < 64 && strings.Contains(intro, " begin") {
		indent += 1
	}
}
