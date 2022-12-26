package httpDevice

import (
	"encoding/xml"
	"fmt"
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
	Username() string
	Password() string
	PollInterval() time.Duration
}

type DeviceStruct struct {
	deviceConfig  device.Config
	teracomConfig Config

	output chan dataflow.Value

	httpClient    *http.Client
	statusRequest *http.Request

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

	c := &DeviceStruct{
		deviceConfig:  deviceConfig,
		teracomConfig: teracomConfig,
		output:        output,

		httpClient: &http.Client{
			// this tool is designed to serve cameras running on the local network
			// -> us a relatively short timeout
			Timeout: 10 * time.Second,
		},

		registers: make(map[string]dataflow.Register),
		sort:      make(map[string]int),
		shutdown:  make(chan struct{}),
	}

	// setup request
	if c.statusRequest, err = c.getStatusRequest(); err != nil {
		return nil, err
	}

	// start polling the device
	c.startPolling()

	return c, nil
}

func (c *DeviceStruct) getRegisterSort(category string) int {
	offset := getCategorySort(category) * 100
	if count, ok := c.sort[category]; !ok {
		c.sort[category] = 1
		return offset
	} else {
		c.sort[category] += 1
		return offset + count
	}
}

func (c *DeviceStruct) startPolling() {
	if c.deviceConfig.LogDebug() {
		log.Printf("httpDevice[%s]: start polling, interval=%s", c.deviceConfig.Name(), c.teracomConfig.PollInterval())
	}

	// start source go routine
	go func() {
		defer close(c.output)

		ticker := time.NewTicker(c.teracomConfig.PollInterval())

		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				statusXml, err := c.getStatusXml()
				if err != nil {
					log.Printf("httpDevice[%s]: canot get status, err=%s", c.deviceConfig.Name(), err)
					continue
				}

				var status StatusStruct
				err = xml.Unmarshal(statusXml, &status)
				if err != nil {
					log.Printf("httpDevice[%s]: canot parse xml, err=%s", c.deviceConfig.Name(), err)
					continue
				}
				c.extractRegistersAndValues(status)
				c.SetLastUpdatedNow()
			}
		}
	}()
}

func (c *DeviceStruct) getStatusRequest() (request *http.Request, err error) {
	addr := c.teracomConfig.Url().JoinPath("status.xml")
	request, err = http.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return
	}
	request.SetBasicAuth(c.teracomConfig.Username(), c.teracomConfig.Password())

	return
}

func (c *DeviceStruct) getStatusXml() (body []byte, err error) {
	resp, err := c.httpClient.Do(c.statusRequest)
	if err != nil {
		log.Printf("httpDevice[%s]: cannot get status: %s", c.deviceConfig.Name(), err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("GET %s failed with code: %d", c.statusRequest.URL.String(), resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	return
}

func (c *DeviceStruct) Config() device.Config {
	return c.deviceConfig
}

func (c *DeviceStruct) ShutdownChan() chan struct{} {
	return c.shutdown
}

func (c *DeviceStruct) addIgnoreRegister(category, registerName, description, unit, dataType string) dataflow.Register {
	// check if this register exists already and the properties are still the same
	c.registersMutex.RLock()
	if r, ok := c.registers[registerName]; ok {
		if r.Category() == category &&
			r.Description() == description &&
			r.Unit() == unit {
			c.registersMutex.RUnlock()
			return r
		}
	}
	c.registersMutex.RUnlock()

	// check if register is on ignore list
	if device.IsExcluded(registerName, category, c.deviceConfig) {
		return nil
	}

	// create new register
	sort := c.getRegisterSort(category)
	var r dataflow.Register
	if dataType == "numeric" {
		r = dataflow.CreateNumberRegisterStruct(
			category,
			registerName,
			description,
			0,
			false,
			true,
			1,
			unit,
			sort,
		)
	} else if dataType == "text" {
		r = dataflow.CreateTextRegisterStruct(
			category,
			registerName,
			description,
			0,
			false,
			sort,
		)
	} else {
		panic("unknown dataType: " + dataType)
	}

	// add the register into the list
	c.registersMutex.Lock()
	defer c.registersMutex.Unlock()

	c.registers[registerName] = r
	return r
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()

	ret := make(dataflow.Registers, len(c.registers))
	i := 0
	for _, r := range c.registers {
		ret[i] = r
		i += 1
	}
	return ret
}

func (c *DeviceStruct) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *DeviceStruct) LastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *DeviceStruct) Model() string {
	return "teracom"
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	log.Printf("httpDevice[%s]: shutdown completed", c.deviceConfig.Name())
}
