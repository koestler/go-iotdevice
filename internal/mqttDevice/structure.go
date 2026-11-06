package mqttDevice

import (
	"cmp"
	"encoding/json"
	"slices"

	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
)

func parseStructPayload(payload []byte) (msg mqttForwarders.StructureMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	slices.Sort(msg.AvailabilityTopics)
	slices.SortFunc(msg.Registers, func(a, b mqttForwarders.StructRegister) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return
}

func equalStructure(a, b mqttForwarders.StructureMessage) bool {
	return !slices.Equal(a.AvailabilityTopics, b.AvailabilityTopics) ||
		a.TelemetryTopic != b.TelemetryTopic ||
		a.RealtimeTopic != b.RealtimeTopic ||
		a.CommandTopic != b.CommandTopic ||
		!slices.Equal(structureRegisterNames(a.Registers), structureRegisterNames(b.Registers))
}

func structureRegisterNames(registers []mqttForwarders.StructRegister) []string {
	names := make([]string, 0, len(registers))
	for _, r := range registers {
		names = append(names, r.Name)
	}
	slices.Sort(names)
	return names
}
