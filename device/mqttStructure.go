package device

import (
	"context"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"golang.org/x/exp/maps"
	"log"
	"math"
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
	Controllable bool           `json:"Cont" example:"false"`
}

type StructureMessage struct {
	AvailabilityTopics []string
	RealtimeTopic      string
	Registers          []StructRegister
}

func runStructureForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
) {
	devCfg := dev.Config()

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(devCfg.ViaMqttClients()) {
		mcCfg := mc.Config()
		if !mcCfg.StructureEnabled() {
			continue
		}

		go func(mc mqttClient.Client) {
			if devCfg.LogDebug() {
				defer func() {
					log.Printf(
						"device[%s]->mqttClient[%s]->struct: exit",
						devCfg.Name(), mcCfg.Name(),
					)
				}()
			}

			// for initial send: wait until first register is received; than wait for 1s
			regSubscription := dev.RegisterDb().Subscribe(ctx)

			// compute topics
			realtimeTopic := mcCfg.RealtimeTopic(devCfg.Name(), "%RegisterName%")
			structureTopic := mcCfg.StructureTopic(devCfg.Name())

			ticker := time.NewTicker(math.MaxInt64)
			defer ticker.Stop()

			registers := make(map[string]StructRegister)
			for {
				select {
				case <-ctx.Done():
					if devCfg.LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]->structure: exit",
							devCfg.Name(), mcCfg.Name(),
						)
					}
					return
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

					publishStruct(mc, devCfg, structureTopic, StructureMessage{
						AvailabilityTopics: getAvailabilityTopics(devCfg, mcCfg),
						RealtimeTopic:      realtimeTopic,
						Registers:          maps.Values(registers),
					})
					registers = make(map[string]StructRegister)
				}
			}
		}(mc)
	}
}

func getAvailabilityTopics(devCfg Config, mcCfg mqttClient.Config) (ret []string) {
	if mcCfg.AvailabilityClientEnabled() {
		ret = append(ret, mcCfg.AvailabilityClientTopic())
	}
	if mcCfg.AvailabilityDeviceEnabled() && stringContains(mcCfg.Name(), devCfg.ViaMqttClients()) {
		ret = append(ret, mcCfg.AvailabilityDeviceTopic(devCfg.Name()))
	}

	return
}

func publishStruct(mc mqttClient.Client, devCfg Config, topic string, msg StructureMessage) {
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
			topic,
			payload,
			mcCfg.Qos(),
			mcCfg.StructureRetain(),
		)
	}
}

func stringContains(needle string, haystack []string) bool {
	for _, t := range haystack {
		if t == needle {
			return true
		}
	}
	return false
}
