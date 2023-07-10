package httpDevice

import (
	"context"
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
}

type DeviceStruct struct {
	device.State
	httpConfig Config

	commandStorage *dataflow.ValueStorageInstance

	httpClient  *http.Client
	pollRequest *http.Request
	impl        Implementation

	registers      map[string]dataflow.Register
	sort           map[string]int
	registersMutex sync.RWMutex
}

func CreateDevice(
	deviceConfig device.Config,
	teracomConfig Config,
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) *DeviceStruct {
	ds := &DeviceStruct{
		State: device.CreateState(
			deviceConfig,
			stateStorage,
		),
		httpConfig:     teracomConfig,
		commandStorage: commandStorage,

		httpClient: &http.Client{
			// this tool is designed to serve devices running on the local network
			// -> us a relatively short timeout
			Timeout: time.Second,
		},

		registers: make(map[string]dataflow.Register),
		sort:      make(map[string]int),
	}

	// setup impl
	ds.impl = implementationFactory(ds)

	return ds
}

func (ds *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	// setup request
	if ds.pollRequest, err = ds.GetRequest(ds.impl.GetPath()); err != nil {
		return err, true
	}

	if ds.Config().LogDebug() {
		log.Printf("httpDevice[%s]: start polling, interval=%s", ds.Name(), ds.httpConfig.PollInterval())
	}

	execPoll := func() error {
		if err := ds.poll(); err != nil {
			return fmt.Errorf("httpDevice[%s]: error: %s", ds.Name(), err)
		} else {
			if ds.Config().LogDebug() {
				log.Printf("httpDevice[%s]: poll request successful", ds.Config().Name())
			}
		}
		return nil
	}
	if err := execPoll(); err != nil {
		return err, true
	}

	// send connected now, disconnected when this routine stops
	ds.SetAvailable(true)
	defer func() {
		ds.SetAvailable(false)
	}()

	// setup subscription to listen for updates of controllable registers
	filter := dataflow.Filter{
		SkipNull:       true,
		IncludeDevices: map[string]bool{ds.Config().Name(): true},
	}
	commandSubscription := ds.commandStorage.Subscribe(filter)
	defer commandSubscription.Shutdown()

	execCommand := func(value dataflow.Value) {
		if ds.Config().LogDebug() {
			log.Printf(
				"httpDevice[%s]: controllable command: %s",
				ds.Name(), value.String(),
			)
		}
		if request, onSuccess, err := ds.impl.ControlValueRequest(value); err != nil {
			log.Printf(
				"httpDevice[%s]: control request genration failed: %s",
				ds.Name(), err,
			)
		} else {
			request.URL = ds.httpConfig.Url().JoinPath(request.URL.String())
			request.SetBasicAuth(ds.httpConfig.Username(), ds.httpConfig.Password())
			if resp, err := ds.httpClient.Do(request); err != nil {
				log.Printf(
					"httpDevice[%s]: control request failed: %s",
					ds.Name(), err,
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

		// reset the command; this allows the same command (eg. toggle) to be sent again
		ds.commandStorage.Fill(dataflow.NewNullRegisterValue(ds.Config().Name(), value.Register()))
	}

	pollTicker := time.NewTicker(ds.httpConfig.PollInterval())
	defer pollTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-pollTicker.C:
			if err := execPoll(); err != nil {
				return err, false
			}
		case value := <-commandSubscription.GetOutput():
			execCommand(value)
		}
	}
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
	if device.IsExcluded(registerName, category, ds.Config()) {
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

	ret := make(dataflow.Registers, len(ds.registers)+1)
	i := 0
	for _, r := range ds.registers {
		ret[i] = r
		i += 1
	}
	ret[len(ds.registers)] = device.GetAvailabilityRegister()
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

func (ds *DeviceStruct) Model() string {
	return ds.httpConfig.Kind().String()
}
