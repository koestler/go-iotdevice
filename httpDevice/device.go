package httpDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Config interface {
	Url() *url.URL
	Kind() config.HttpDeviceKind
	Username() string
	Password() string
	PollInterval() time.Duration
	PollIntervalMaxBackoff() time.Duration
}

type DeviceStruct struct {
	deviceConfig device.Config
	httpConfig   Config

	output chan dataflow.Value

	httpClient  *http.Client
	pollRequest *http.Request
	impl        Implementation

	registers        map[string]dataflow.Register
	sort             map[string]int
	registersMutex   sync.RWMutex
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex

	shutdown chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	teracomConfig Config,
	storage *dataflow.ValueStorageInstance,
	mqttClientPool *mqttClient.ClientPool,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)

	// create source and connect to storage
	source := dataflow.CreateSource(output)
	source.Append(storage)

	ds := &DeviceStruct{
		deviceConfig: deviceConfig,
		httpConfig:   teracomConfig,
		output:       output,

		httpClient: &http.Client{
			// this tool is designed to serve cameras running on the local network
			// -> us a relatively short timeout
			Timeout: 10 * time.Second,
		},

		registers: make(map[string]dataflow.Register),
		sort:      make(map[string]int),
		shutdown:  make(chan struct{}),
	}

	// setup impl
	ds.impl = implementationFactory(ds)

	// setup request
	if ds.pollRequest, err = ds.GetRequest(ds.impl.GetPath()); err != nil {
		return
	}

	// start polling the device
	ds.pollingRoutine()

	return ds, nil
}

func (ds *DeviceStruct) GetRequest(path string) (request *http.Request, err error) {
	addr := ds.httpConfig.Url().JoinPath(path)
	request, err = http.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return
	}
	request.SetBasicAuth(ds.httpConfig.Username(), ds.httpConfig.Password())

	return
}

func (ds *DeviceStruct) getRegisterSort(category string) int {
	offset := ds.impl.GetCategorySort(category) * 100
	if count, ok := ds.sort[category]; !ok {
		ds.sort[category] = 1
		return offset
	} else {
		ds.sort[category] += 1
		return offset + count
	}
}

func (ds *DeviceStruct) pollingRoutine() {
	if ds.deviceConfig.LogDebug() {
		log.Printf("httpDevice[%s]: start polling, interval=%s", ds.deviceConfig.Name(), ds.httpConfig.PollInterval())
	}

	// start source go routine
	errorsInARow := 0
	interval := ds.getPollInterval(errorsInARow)
	go func() {
		defer close(ds.output)
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ds.shutdown:
				return
			case <-ticker.C:
				if err := ds.poll(); err != nil {
					log.Printf("httpDevice[%s]: error: %s", ds.deviceConfig.Name(), err)
					errorsInARow += 1
				} else {
					errorsInARow = 0
				}

				// change poll interval on error
				if newInterval := ds.getPollInterval(errorsInARow); interval != newInterval {
					interval = newInterval
					ticker.Reset(interval)
					if errorsInARow == 0 {
						log.Printf("httpDevice[%s]: recoverd, next poll in: %s", ds.deviceConfig.Name(), interval)
					}
				}
				if errorsInARow > 0 {
					log.Printf("httpDevice[%s]: exponential backoff, retry in: %s", ds.deviceConfig.Name(), interval)
				}
			}
		}
	}()
}

func (ds *DeviceStruct) getPollInterval(errorsInARow int) time.Duration {
	if errorsInARow > 62 {
		errorsInARow = 62
	}
	var backoffFactor uint64 = 1 << errorsInARow // 2^errorsInARow
	interval := ds.httpConfig.PollInterval() * time.Duration(backoffFactor)
	max := ds.httpConfig.PollIntervalMaxBackoff()
	if interval > max {
		return max
	}
	return interval
}

func (ds *DeviceStruct) poll() error {
	resp, err := ds.httpClient.Do(ds.pollRequest)
	if err != nil {
		return fmt.Errorf("cannot get status: %s", err)

	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s failed with code: %d", ds.pollRequest.URL.String(), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("error during body reader close: %s", err)
	}
	if err != nil {
		return fmt.Errorf("cannot get response body: %s", err)
	}

	if err := ds.impl.HandleResponse(body); err != nil {
		return err
	}

	ds.SetLastUpdatedNow()
	return nil
}

func (ds *DeviceStruct) Config() device.Config {
	return ds.deviceConfig
}

func (ds *DeviceStruct) ShutdownChan() chan struct{} {
	return ds.shutdown
}

func (ds *DeviceStruct) addIgnoreRegister(category, registerName, description, unit, dataType string) dataflow.Register {
	// check if this register exists already and the properties are still the same
	ds.registersMutex.RLock()
	if r, ok := ds.registers[registerName]; ok {
		if r.Category() == category &&
			r.Description() == description &&
			r.Unit() == unit {
			ds.registersMutex.RUnlock()
			return r
		}
	}
	ds.registersMutex.RUnlock()

	// check if register is on ignore list
	if device.IsExcluded(registerName, category, ds.deviceConfig) {
		return nil
	}

	// create new register
	sort := ds.getRegisterSort(category)
	var r dataflow.Register
	var registerType dataflow.RegisterType

	if dataType == "numeric" {
		registerType = dataflow.NumberRegister
	} else if dataType == "text" {
		registerType = dataflow.TextRegister
	} else {
		panic("unknown dataType: " + dataType)
	}

	r = dataflow.CreateRegisterStruct(
		category,
		registerName,
		description,
		registerType,
		unit,
		sort,
	)

	// add the register into the list
	ds.registersMutex.Lock()
	defer ds.registersMutex.Unlock()

	ds.registers[registerName] = r
	return r
}

func (ds *DeviceStruct) Registers() dataflow.Registers {
	ds.registersMutex.RLock()
	defer ds.registersMutex.RUnlock()

	ret := make(dataflow.Registers, len(ds.registers))
	i := 0
	for _, r := range ds.registers {
		ret[i] = r
		i += 1
	}
	return ret
}

func (ds *DeviceStruct) SetLastUpdatedNow() {
	ds.lastUpdatedMutex.Lock()
	defer ds.lastUpdatedMutex.Unlock()
	ds.lastUpdated = time.Now()
}

func (ds *DeviceStruct) LastUpdated() time.Time {
	ds.lastUpdatedMutex.RLock()
	defer ds.lastUpdatedMutex.RUnlock()
	return ds.lastUpdated
}

func (ds *DeviceStruct) Model() string {
	return ds.httpConfig.Kind().String()
}

func (ds *DeviceStruct) Shutdown() {
	close(ds.shutdown)
	log.Printf("httpDevice[%s]: shutdown completed", ds.deviceConfig.Name())
}
