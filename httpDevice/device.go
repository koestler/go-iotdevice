package httpDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/types"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Config interface {
	Url() *url.URL
	Kind() types.HttpDeviceKind
	Username() string
	Password() string
	PollInterval() time.Duration
}

type DeviceStruct struct {
	device.State
	httpConfig     Config
	registerFilter dataflow.RegisterFilterFunc

	commandStorage *dataflow.ValueStorage

	httpClient  *http.Client
	pollRequest *http.Request
	impl        Implementation

	sort map[string]int
}

func NewDevice(
	deviceConfig device.Config,
	teracomConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	ds := &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		httpConfig:     teracomConfig,
		registerFilter: dataflow.RegisterFilter(deviceConfig.Filter()),
		commandStorage: commandStorage,

		httpClient: &http.Client{
			// this tool is designed to serve devices running on the local network
			// -> us a relatively short timeout
			Timeout: time.Second,
		},

		sort: make(map[string]int),
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

	// setup subscription to listen for updates of writable registers
	_, commandSubscription := ds.commandStorage.SubscribeReturnInitial(ctx, dataflow.DeviceNonNullValueFilter(ds.Config().Name()))

	execCommand := func(value dataflow.Value) {
		if ds.Config().LogDebug() {
			log.Printf(
				"httpDevice[%s]: command: %s",
				ds.Name(), value.String(),
			)
		}
		if request, onSuccess, err := ds.impl.CommandValueRequest(value); err != nil {
			log.Printf(
				"httpDevice[%s]: command request genration failed: %s",
				ds.Name(), err,
			)
		} else {
			request.URL = ds.httpConfig.Url().JoinPath(request.URL.String())
			request.SetBasicAuth(ds.httpConfig.Username(), ds.httpConfig.Password())
			if resp, err := ds.httpClient.Do(request); err != nil {
				log.Printf(
					"httpDevice[%s]: command request failed: %s",
					ds.Name(), err,
				)
			} else {
				// ready and discard response body
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Printf(
						"httpDevice[%s]: command request failed with code: %d",
						ds.Config().Name(), resp.StatusCode,
					)
					return
				}

				if _, err = io.ReadAll(resp.Body); err != nil {
					log.Printf(
						"httpDevice[%s]: command cannot read body: %s",
						ds.Config().Name(), err,
					)
					return
				}

				if ds.Config().LogDebug() {
					log.Printf("httpDevice[%s]: command request successful", ds.Config().Name())
				}
				onSuccess()
			}
		}

		// reset the command; this allows the same command (e.g. toggle) to be sent again
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
		case value := <-commandSubscription.Drain():
			if value != nil {
				execCommand(value)
			}
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

	return nil
}

func (ds *DeviceStruct) addIgnoreRegister(
	category, registerName, description, unit string,
	registerType dataflow.RegisterType,
	enum map[int]string,
	writable bool,
) dataflow.Register {
	// check if this register exists already and the properties are still the same
	if r, ok := ds.RegisterDb().GetByName(registerName); ok {
		if r.Category() == category &&
			r.Description() == description &&
			r.RegisterType() == registerType &&
			r.Unit() == unit {
			return r
		}
	}

	// create new register
	sort := ds.getRegisterSort(category)
	r := dataflow.NewRegisterStruct(
		category,
		registerName,
		description,
		registerType,
		enum,
		unit,
		sort,
		writable,
	)

	// check if register is on ignore list
	if !ds.registerFilter(r) {
		return nil
	}

	ds.Config().Filter()

	// add the register into the list
	ds.RegisterDb().Add(r)

	return r
}

func (ds *DeviceStruct) Model() string {
	return ds.httpConfig.Kind().String()
}
