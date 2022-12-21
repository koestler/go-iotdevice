package teracomDevice

import (
	"crypto/tls"
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

	source *dataflow.Source

	httpClient    *http.Client
	statusRequest *http.Request

	registers        map[string]dataflow.Register
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
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(storage)

	c := &DeviceStruct{
		deviceConfig:  deviceConfig,
		teracomConfig: teracomConfig,
		source:        source,

		httpClient:  &http.Client{
			// this tool is designed to serve cameras running on the local network
			// -> us a relatively short timeout
			Timeout: 10 * time.Second,

			// ubnt cameras don't use valid certificates
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},

		registers:     make(map[string]dataflow.Register),
		shutdown:      make(chan struct{}),
	}

	// setup request
	if c.statusRequest, err = c.getStatusRequest(); err != nil {
		return nil, err
	}

	// start polling the device
	c.startPolling(output)

	return c, nil
}

func (c *DeviceStruct) startPolling(output chan dataflow.Value) {
	if c.deviceConfig.LogDebug() {
		log.Printf("teracomDevice[%s]: start polling, interval=%s", c.deviceConfig.Name(), c.teracomConfig.PollInterval())
	}

	// start source go routine
	go func() {
		defer close(output)

		ticker := time.NewTicker(c.teracomConfig.PollInterval())

		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				log.Printf("teracomDevice[%s]: tick", c.deviceConfig.Name())

				xml, err := c.getStatusXml()
				log.Printf("teracomDevice[%s]: err=%s", c.deviceConfig.Name(), err)
				log.Printf("teracomDevice[%s]: xml=%s", c.deviceConfig.Name(), xml)

				c.SetLastUpdatedNow()
			}
		}
	}()
}

func (c *DeviceStruct) getStatusRequest() (request *http.Request, err error) {
	addr := c.teracomConfig.Url().JoinPath("status.xml")
	request, err = http.NewRequest("GET", addr.String(), nil)
	if err != nil { return }
	request.SetBasicAuth(c.teracomConfig.Username(), c.teracomConfig.Password())

	return
}

func (c *DeviceStruct) getStatusXml() (body []byte, err error) {
	resp, err := c.httpClient.Do(c.statusRequest)
	if err != nil {
		log.Printf("teracomDevice[%s]: cannot get status: %s", c.deviceConfig.Name(), err)
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
	log.Printf("teracomDevice[%s]: shutdown completed", c.deviceConfig.Name())
}
