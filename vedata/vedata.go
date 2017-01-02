package vedata

import (
	"errors"
	"github.com/koestler/go-ve-sensor/bmv"
	"log"
)

type DeviceId uint64

type Device struct {
	Name          string
	NumericValues bmv.NumericValues
}

type readDeviceOp struct {
	deviceId DeviceId
	response chan bool
	err      error
	device   Device
}

type readDeviceIdsOp struct {
	response chan []DeviceId
}

type writeOp struct {
	deviceId      DeviceId
	numericValues bmv.NumericValues
	response      chan bool
}

type DbType map[DeviceId]Device

var running bool
var db DbType
var readDeviceChan chan *readDeviceOp
var readDeviceIdsChan chan *readDeviceIdsOp
var writes chan *writeOp

func init() {
	running = false
	db = make(map[DeviceId]Device)

	readDeviceChan = make(chan *readDeviceOp)
	readDeviceIdsChan = make(chan *readDeviceIdsOp)
	writes = make(chan *writeOp)
}

func CreateDevice(name string) (deviceId DeviceId) {
	if running {
		log.Panic("must no call vedata.CreateDevice after vedata.Run")
	}

	deviceId = DeviceId(len(db) + 1)

	db[deviceId] = Device{
		Name:          name,
		NumericValues: make(bmv.NumericValues),
	}

	return
}

func (deviceId DeviceId) ReadDevice() (device Device, err error) {
	read := &readDeviceOp{
		deviceId: deviceId,
		response: make(chan bool),
		err:      nil,
		device:   Device{},
	}
	readDeviceChan <- read
	<-read.response

	return read.device, read.err
}

func ReadDeviceIds() (ret []DeviceId) {
	read := &readDeviceIdsOp{
		response: make(chan []DeviceId)}
	readDeviceIdsChan <- read
	ret = <-read.response
	return
}

func (deviceId DeviceId) Write(numericValues bmv.NumericValues) {
	write := &writeOp{
		deviceId:      deviceId,
		numericValues: numericValues,
		response:      make(chan bool),
	}
	writes <- write
	<-write.response
}

func Run() {
	go func() {
		running = true
		for {
			select {
			case write := <-writes:
				for k, v := range write.numericValues {
					db[write.deviceId].NumericValues[k] = v
				}
				write.response <- true
			case read := <-readDeviceChan:
				var ok bool
				read.device, ok = db[read.deviceId]
				if !ok {
					read.err = errors.New("device not found")
				}
				read.response <- true
			case read := <-readDeviceIdsChan:
				deviceIds := make([]DeviceId, len(db))
				i := 0
				for k, _ := range db {
					deviceIds[i] = k
					i++
				}
				read.response <- deviceIds
			}
		}
	}()
}
