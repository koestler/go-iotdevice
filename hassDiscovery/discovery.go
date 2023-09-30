package hassDiscovery

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
)

// discoveryPrefix: defined in the HassDiscovery config section
// component: sensor / switch
// nodeId: use the ClientId as defined in the MqttClients config section
// objectId: use deviceName-registerName
func getTopic(discoveryPrefix, component, nodeId, objectId string) string {
	return fmt.Sprintf("%s/%s/%s/%s/config", discoveryPrefix, component, nodeId, objectId)
}

type availabilityStruct struct {
	Topic string `json:"t"`
}

type discoveryMessage struct {
	UniqueId string `json:"uniq_id"`
	Name     string `json:"name"`

	StateTopic        string               `json:"stat_t"`
	Availability      []availabilityStruct `json:"avty"`
	AvailabilityMode  string               `json:"avty_mode"`
	ValueTemplate     string               `json:"val_tpl"`
	UnitOfMeasurement string               `json:"unit_of_meas,omitempty"`
}

func getSensorMessage(
	discoveryPrefix string,
	mcCfg mqttClient.Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate string,
) (topic string, msg discoveryMessage) {
	uniqueId := fmt.Sprintf("%s-%s", deviceName, CamelToSnakeCase(register.Name()))
	name := fmt.Sprintf("%s %s", deviceName, register.Description())

	topic = getTopic(discoveryPrefix, "sensor", mcCfg.ClientId(), uniqueId)

	msg = discoveryMessage{
		UniqueId:          uniqueId,
		Name:              name,
		StateTopic:        mcCfg.RealtimeTopic(deviceName, register.Name()),
		Availability:      getAvailabilityTopics(deviceName, mcCfg),
		AvailabilityMode:  "all",
		ValueTemplate:     valueTemplate,
		UnitOfMeasurement: register.Unit(),
	}

	return
}

func getAvailabilityTopics(deviceName string, mcCfg mqttClient.Config) (ret []availabilityStruct) {
	if mcCfg.AvailabilityClientEnabled() {
		ret = append(ret, availabilityStruct{mcCfg.AvailabilityClientTopic()})
	}
	if mcCfg.AvailabilityDeviceEnabled() {
		ret = append(ret, availabilityStruct{mcCfg.AvailabilityDeviceTopic(deviceName)})
	}

	return
}
