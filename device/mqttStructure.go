package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
)

type StructRegister struct {
	Category     string         `json:"Cat" example:"Monitor"`
	Name         string         `json:"Name" example:"PanelPower"`
	Description  string         `json:"Desc" example:"Panel power"`
	Type         string         `json:"Type" example:"numeric"`
	Enum         map[int]string `json:"Enum,omitempty"`
	Unit         string         `json:"Unit,omitempty" example:"W"`
	Sort         int            `json:"Sort" example:"100"`
	Controllable *bool          `json:"Control,omitempty" example:"false"`
}

type StructureMessage struct {
	AvailabilityTopic string
	RealtimeTopic     string
	Registers         []StructRegister
}

func runStructureForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	devCfg := dev.Config()

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(dev.Config().RealtimeViaMqttClients()) {
		mcCfg := mc.Config()
		if !mcCfg.StructureEnabled() {
			continue
		}

		registers := make(map[string]StructRegister)

		go func(mc mqttClient.Client) {
			if dev.Config().LogDebug() {
				defer func() {
					log.Printf(
						"device[%s]->mqttClient[%s]->struct: exit",
						dev.Config().Name(), mcCfg.Name(),
					)
				}()
			}

			// for initial send: wait until first register is received; than wait until no new register is found for 1s
			regSubscription := dev.RegisterDb().Subscribe(ctx)

			for {
				select {}

			}

		}(mc)
	}
}
