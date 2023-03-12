package httpDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
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
	deviceConfig   device.Config
	httpConfig     Config
	stateStorage   *dataflow.ValueStorageInstance
	commandStorage *dataflow.ValueStorageInstance

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
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)

	// create source and connect to storage
	source := dataflow.CreateSource(output)
	source.Append(stateStorage)

	ds := &DeviceStruct{
		deviceConfig:   deviceConfig,
		httpConfig:     teracomConfig,
		stateStorage:   stateStorage,
		commandStorage: commandStorage,
		output:         output,

		httpClient: &http.Client{
			// this tool is designed to serve devices running on the local network
			// -> us a relatively short timeout
			Timeout: time.Second,
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

	ds.mainRoutine()

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

func (ds *DeviceStruct) mainRoutine() {
	if ds.deviceConfig.LogDebug() {
		log.Printf("httpDevice[%s]: start polling, interval=%s", ds.deviceConfig.Name(), ds.httpConfig.PollInterval())
	}

	go func() {
		// setup polling interval and logging
		errorsInARow := 0
		lastErrorMsg := ""
		interval := ds.getPollInterval(errorsInARow)

		defer close(ds.output)
		pollTicker := time.NewTicker(interval)

		ds.addRegister(device.GetAvailabilityRegister())
		execPoll := func() {
			if err := ds.poll(); err != nil {
				errorsInARow += 1
				if errMsg := fmt.Sprintf("httpDevice[%s]: error: %s", ds.deviceConfig.Name(), err); errMsg != lastErrorMsg {
					log.Println(errMsg)
					lastErrorMsg = errMsg
				}
				if errorsInARow > 1 {
					device.SendDisconnected(ds.Config().Name(), ds.output)
				}
			} else {
				device.SendConnteced(ds.Config().Name(), ds.output)
				errorsInARow = 0
				lastErrorMsg = ""
				if ds.Config().LogDebug() {
					log.Printf("httpDevice[%s]: poll request successful", ds.Config().Name())
				}
			}

			// change poll interval on error
			if newInterval := ds.getPollInterval(errorsInARow); interval != newInterval {
				interval = newInterval
				pollTicker.Reset(interval)
				if errorsInARow == 0 {
					log.Printf("httpDevice[%s]: recoverd, next poll in: %s", ds.deviceConfig.Name(), interval)
				} else {
					log.Printf("httpDevice[%s]: exponential backoff, retry in: %s", ds.deviceConfig.Name(), interval)
				}
			}
		}
		execPoll()

		// setup subscription to listen for updates of controllable registers
		filter := dataflow.Filter{
			IncludeDevices: map[string]bool{ds.Config().Name(): true},
		}
		commandSubscription := ds.commandStorage.Subscribe(filter)
		defer commandSubscription.Shutdown()

		execCommand := func(value dataflow.Value) {
			if ds.Config().LogDebug() {
				log.Printf(
					"httpDevice[%s]: controllable command: %s",
					ds.Config().Name(), value.String(),
				)
			}
			if request, onSuccess, err := ds.impl.ControlValueRequest(value); err != nil {
				log.Printf(
					"httpDevice[%s]: control request genration failed: %s",
					ds.Config().Name(), err,
				)
			} else {
				request.URL = ds.httpConfig.Url().JoinPath(request.URL.String())
				request.SetBasicAuth(ds.httpConfig.Username(), ds.httpConfig.Password())
				if resp, err := ds.httpClient.Do(request); err != nil {
					log.Printf(
						"httpDevice[%s]: control request failed: %s",
						ds.Config().Name(), err,
					)
				} else {
					// ready and discard response body
					defer resp.Body.Close()
					io.ReadAll(resp.Body)

					if resp.StatusCode != http.StatusOK {
						log.Printf(
							"httpDevice[%s]: control request failed with code: %d",
							ds.Config().Name(), resp.StatusCode,
						)
					} else {
						if ds.Config().LogDebug() {
							log.Printf("httpDevice[%s]: control request successful", ds.Config().Name())
						}
						onSuccess()
					}
				}
			}
		}

		for {
			select {
			case <-ds.shutdown:
				return
			case <-pollTicker.C:
				execPoll()
			case value := <-commandSubscription.GetOutput():
				execCommand(value)
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
		return fmt.Errorf("GET %s failed", ds.pollRequest.URL.String())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot get response body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s failed with code: %d", ds.pollRequest.URL.String(), resp.StatusCode)
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

	ds.addRegister(r)

	return r
}

func (ds *DeviceStruct) addRegister(register dataflow.Register) {
	ds.registersMutex.Lock()
	defer ds.registersMutex.Unlock()

	ds.registers[register.Name()] = register
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

func (ds *DeviceStruct) GetRegister(registerName string) dataflow.Register {
	ds.registersMutex.RLock()
	defer ds.registersMutex.RUnlock()

	if r, ok := ds.registers[registerName]; ok {
		return r
	}
	return nil
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
