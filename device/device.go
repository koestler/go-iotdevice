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
	Registers() []dataflow.Register
	GetRegister(registerName string) dataflow.Register
	LastUpdated() time.Time
	IsAvailable() bool
	Model() string
	Run(ctx context.Context) (err error, immediateError bool)
}

type State struct {
	deviceConfig Config
	stateStorage *dataflow.ValueStorageInstance

	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	available        bool
	availableMutex   sync.RWMutex
}

func CreateState(deviceConfig Config, stateStorage *dataflow.ValueStorageInstance) State {
	return State{
		deviceConfig: deviceConfig,
		stateStorage: stateStorage,
	}
}

func (c *State) Name() string {
	return c.deviceConfig.Name()
}

func (c *State) Config() Config {
	return c.deviceConfig
}

func (c *State) StateStorage() *dataflow.ValueStorageInstance {
	return c.stateStorage
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
