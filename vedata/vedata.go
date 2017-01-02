package vedata

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"log"
)

type DeviceId uint64

type Device struct {
	Name          string
	NumericValues bmv.NumericValues
}

type readOp struct {
	deviceId DeviceId
	response chan bmv.NumericValues
}

type writeOp struct {
	deviceId DeviceId
	key      string
	value    bmv.NumericValue
	response chan bool
}

var db map[DeviceId]Device

var reads chan *readOp
var writes chan *writeOp

func init() {
	db = make(map[DeviceId]Device)

	reads = make(chan *readOp)
	writes = make(chan *writeOp)
}

func CreateDevice(name string) (deviceId DeviceId) {
	deviceId = DeviceId(len(db) + 1)

	db[deviceId] = Device{
		Name:          name,
		NumericValues: make(bmv.NumericValues),
	}

	return
}

func (deviceId DeviceId) Read() {
	read := &readOp{
		deviceId: deviceId,
		response: make(chan bmv.NumericValues)}
	reads <- read
	<-read.response
}

func (deviceId DeviceId) WriteNumericValue(key string, value bmv.NumericValue) {
	write := &writeOp{
		deviceId: deviceId,
		key:      key,
		value:    value,
		response: make(chan bool),
	}
	writes <- write
	<-write.response
}

func Run() {
	go func() {
		for {
			select {
			case write := <-writes:
				db[write.deviceId].NumericValues[write.key] = write.value
				write.response <- true
			case read := <-reads:
				log.Printf("vedata.Read %v", read)
				read.response <- db[read.deviceId].NumericValues
			}
		}
	}()
}
