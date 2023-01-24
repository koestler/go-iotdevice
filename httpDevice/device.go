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
	lastErrorMsg := ""
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
					errorsInARow += 1
					if errMsg := fmt.Sprintf("httpDevice[%s]: error: %s", ds.deviceConfig.Name(), err); errMsg != lastErrorMsg {
						log.Println(errMsg)
						lastErrorMsg = errMsg
					}
				} else {
					errorsInARow = 0
					lastErrorMsg = ""
				}

				// change poll interval on error
				if newInterval := ds.getPollInterval(errorsInARow); interval != newInterval {
					interval = newInterval
					ticker.Reset(interval)
					if errorsInARow == 0 {
						log.Printf("httpDevice[%s]: recoverd, next poll in: %s", ds.deviceConfig.Name(), interval)
					} else {
						log.Printf("httpDevice[%s]: exponential backoff, retry in: %s", ds.deviceConfig.Name(), interval)
					}
				}
			}
		}
	}()
}

func (ds *DeviceStruct) getPollInterval(errorsInARow int) time.Duration {
	if errorsInARow > 16 {
		errorsInARow = 16
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
		return fmt.Errorf("GET %s failed: %d", ds.pollRequest.URL.String())
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

func (ds *DeviceStruct) addIgnoreRegister(
	category, registerName, description, unit string,
	registerType dataflow.RegisterType,
	enum map[int]string,
	controllable bool,
) dataflow.Register {
	// check if this register exists already and the properties are still the same
	ds.registersMutex.RLock()
	if r, ok := ds.registers[registerName]; ok {
		if r.Category() == category &&
			r.Description() == description &&
			r.RegisterType() == registerType &&
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
	r := dataflow.CreateRegisterStruct(
		category,
		registerName,
		description,
		registerType,
		enum,
		unit,
		sort,
		controllable,
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
