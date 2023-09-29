package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"golang.org/x/exp/maps"
	"log"
	"math"
	"strings"
	"time"
)

type StructRegister struct {
	Category     string         `json:"Cat" example:"Monitor"`
	Name         string         `json:"Name" example:"PanelPower"`
	Description  string         `json:"Desc" example:"Panel power"`
	Type         string         `json:"Type" example:"number"`
	Enum         map[int]string `json:"Enum,omitempty"`
	Unit         string         `json:"Unit,omitempty" example:"W"`
	Sort         int            `json:"Sort" example:"100"`
	Controllable bool           `json:"Control" example:"false"`
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

		go func(mc mqttClient.Client) {
			if devCfg.LogDebug() {
				defer func() {
					log.Printf(
						"device[%s]->mqttClient[%s]->struct: exit",
						dev.Config().Name(), mcCfg.Name(),
					)
				}()
			}

			// for initial send: wait until first register is received; than wait for 1s
			regSubscription := dev.RegisterDb().Subscribe(ctx)

			ticker := time.NewTicker(math.MaxInt64)
			defer ticker.Stop()

			registers := make(map[string]StructRegister)
			for {
				select {
				case reg := <-regSubscription:
					if len(registers) < 1 {
						ticker.Reset(time.Second)
					}
					registers[reg.Name()] = StructRegister{
						Category:     reg.Category(),
						Name:         reg.Name(),
						Description:  reg.Description(),
						Type:         reg.RegisterType().String(),
						Enum:         reg.Enum(),
						Unit:         reg.Unit(),
						Sort:         reg.Sort(),
						Controllable: reg.Controllable(),
					}
				case <-ticker.C:
					ticker.Stop()

					publishStruct(mc, devCfg, StructureMessage{
						AvailabilityTopic: mcCfg.AvailabilityTopic(),
						RealtimeTopic:     mcCfg.RealtimeTopic(),
						Registers:         maps.Values(registers),
					})
					registers = make(map[string]StructRegister)
				}
			}
		}(mc)
	}
}

func getStructureTopic(
	topic string,
	deviceName string,
) string {
	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	return topic
}

func publishStruct(mc mqttClient.Client, devCfg Config, msg StructureMessage) {
	mcCfg := mc.Config()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->struct: send: %v",
			devCfg.Name(), mcCfg.Name(), msg,
		)
	}

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf(
			"device[%s]->mqttClient[%s]->struct: cannot generate message: %s",
			devCfg.Name(), mcCfg.Name(), err,
		)
	} else {
		mc.Publish(
			getStructureTopic(mcCfg.StructureTopic(), devCfg.Name()),
			payload,
			mcCfg.Qos(),
			mcCfg.StructureRetain(),
		)
	}
}
