package hassDiscovery

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
)

// discoveryPrefix: defined in the HassDiscovery config section
// component: sensor / switch
// nodeId: use the ClientId as defined in the MqttClients config section
// objectId: use deviceName-registerName
func getTopic(discoveryPrefix, component, nodeId, objectId string) string {
	return fmt.Sprintf("%s/%s/%s/%s/config", discoveryPrefix, component, nodeId, objectId)
}

type discoveryMessage struct {
	UniqueId string `json:"uniq_id"`
	Name     string `json:"name"`

	StateTopic        string `json:"stat_t"`
	AvailabilityTopic string `json:"avty_t"`
	ValueTemplate     string `json:"val_tpl"`
	UnitOfMeasurement string `json:"unit_of_meas,omitempty"`
}

func getSensorMessage(
	discoveryPrefix string,
	mCfg mqttClient.Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate string,
) (topic string, msg discoveryMessage) {
	uniqueId := fmt.Sprintf("%s-%s", deviceName, CamelToSnakeCase(register.Name()))
	name := fmt.Sprintf("%s %s", deviceName, register.Description())

	topic = getTopic(discoveryPrefix, "sensor", mCfg.ClientId(), uniqueId)

	msg = discoveryMessage{
		UniqueId: uniqueId,
		Name:     name,
		StateTopic: mqttClient.ReplaceTemplate(device.GetRealtimeTopic(
			mCfg.RealtimeTopic(),
			deviceName,
			register,
		),
			mCfg,
		),
		AvailabilityTopic: mqttClient.ReplaceTemplate(
			mCfg.AvailabilityTopic(),
			mCfg,
		),
		ValueTemplate:     valueTemplate,
		UnitOfMeasurement: register.Unit(),
	}

	return
}
