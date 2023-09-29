package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"sync"
	"time"
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
	LastUpdated() time.Time
	IsAvailable() bool
	Model() string
	Run(ctx context.Context) (err error, immediateError bool)
}

type State struct {
	deviceConfig Config
	stateStorage *dataflow.ValueStorage
	registerDb   *dataflow.RegisterDb

	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	available        bool
	availableMutex   sync.RWMutex
}

func NewState(deviceConfig Config, stateStorage *dataflow.ValueStorage) State {
	registerDb := dataflow.NewRegisterDb()
	registerDb.Add(availabilityRegister)
	return State{
		deviceConfig: deviceConfig,
		stateStorage: stateStorage,
		registerDb:   registerDb,
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

func (c *State) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *State) LastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *State) SetAvailable(available bool) {
	c.availableMutex.Lock()
	defer c.availableMutex.Unlock()
	c.available = available
	if available {
		c.stateStorage.Fill(dataflow.NewEnumRegisterValue(c.deviceConfig.Name(), availabilityRegister, 1))
	} else {
		c.stateStorage.Fill(dataflow.NewEnumRegisterValue(c.deviceConfig.Name(), availabilityRegister, 0))
	}
}

func (c *State) IsAvailable() bool {
	c.availableMutex.RLock()
	defer c.availableMutex.RUnlock()
	return c.available
}
