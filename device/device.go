package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
)

type Config interface {
	Name() string
	SkipFields() []string
	SkipCategories() []string
	TelemetryViaMqttClients() []string
	RealtimeViaMqttClients() []string
	LogDebug() bool
	LogComDebug() bool
}

type Device interface {
	Name() string
	Config() Config
	RegisterDb() *dataflow.RegisterDb
	IsAvailable() bool
	Model() string
	Run(ctx context.Context) (err error, immediateError bool)
}

type State struct {
	deviceConfig Config
	stateStorage *dataflow.ValueStorage
	registerDb   *dataflow.RegisterDb

	unavailableValue dataflow.Value
	availableValue   dataflow.Value
}

func NewState(deviceConfig Config, stateStorage *dataflow.ValueStorage) State {
	registerDb := dataflow.NewRegisterDb()
	registerDb.Add(availabilityRegister)
	return State{
		deviceConfig: deviceConfig,
		stateStorage: stateStorage,
		registerDb:   registerDb,

		unavailableValue: dataflow.NewEnumRegisterValue(deviceConfig.Name(), availabilityRegister, 0),
		availableValue:   dataflow.NewEnumRegisterValue(deviceConfig.Name(), availabilityRegister, 1),
	}
}

func (c *State) Name() string {
	return c.deviceConfig.Name()
}

func (c *State) Config() Config {
	return c.deviceConfig
}

func (c *State) StateStorage() *dataflow.ValueStorage {
	return c.stateStorage
}

func (c *State) RegisterDb() *dataflow.RegisterDb {
	return c.registerDb
}

func (c *State) SetAvailable(v bool) {
	if v {
		c.stateStorage.Fill(c.availableValue)
	} else {
		c.stateStorage.Fill(c.unavailableValue)
	}
}

func (c *State) IsAvailable() bool {
	state := c.stateStorage.GetStateFiltered(availabilityValueFilter)
	return len(state) > 0 && state[0].Equals(c.availableValue)
}

func (c *State) SubscribeAvailable(ctx context.Context) (initial bool, subscription dataflow.ValueSubscription) {
	initialState, subscription := c.stateStorage.SubscribeReturnInitial(ctx, availabilityValueFilter)
	initial = len(initialState) > 0 && initialState[0].Equals(c.availableValue)
	return
}
