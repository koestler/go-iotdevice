package vedirect

import (
	"log"
	"strings"
	"fmt"
)

var indent = 0

func debugPrintf(format string, v ...interface{}) {
	intro := strings.Split(format, "=")[0]

	if indent > 0 && strings.Contains(intro, "end") {
		indent -= 1
	}

	s := fmt.Sprintf(format, v...)
	s = strings.Replace(s, "\n", "\\n", -1)

	log.Print(strings.Repeat("  ", indent) + s)

	if indent < 64 && strings.Contains(intro, "begin") {
		indent += 1
	}
}
