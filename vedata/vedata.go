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
	deviceId      DeviceId
	numericValues bmv.NumericValues
	response      chan bool
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
		for {
			select {
			case write := <-writes:
				for k, v := range write.numericValues {
					db[write.deviceId].NumericValues[k] = v
				}
				write.response <- true
			case read := <-reads:
				log.Printf("vedata.Read %v", read)
				read.response <- db[read.deviceId].NumericValues
			}
		}
	}()
}
