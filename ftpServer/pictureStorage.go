package ftpServer

import (
	"github.com/koestler/go-ve-sensor/storage"
	"errors"
	"time"
)

var PictureStorage = PictureStorageCreate()

type Picture struct {
	created time.Time
	path    string
}

type State map[*storage.Device]*Picture

type writeRequest struct {
	device  *storage.Device
	picture *Picture
}

type readRequest struct {
	device   *storage.Device
	picture  *Picture
	response chan error
}

type PictureStorageInstance struct {
	state State

	// communication channels to/from the main go routine
	writeRequestChannel chan writeRequest
	readRequestChannel  chan readRequest
}

func (instance *PictureStorageInstance) mainStorageRoutine() {
	for {
		select {
		case writeRequest := <-instance.writeRequestChannel:
			instance.state[writeRequest.device] = writeRequest.picture
		case readRequest := <-instance.readRequestChannel:
			if picture, ok := instance.state[readRequest.device]; ok {
				readRequest.picture = picture
				readRequest.response <- nil
			} else {
				readRequest.response <- errors.New("device does not exist")
			}
		}
	}
}

func PictureStorageCreate() (instance *PictureStorageInstance) {
	instance = &PictureStorageInstance{
		state:               make(State),
		writeRequestChannel: make(chan writeRequest, 4), // input channel is buffered
		readRequestChannel:  make(chan readRequest),
	}

	// main go routine
	go instance.mainStorageRoutine()

	return
}

func (instance *PictureStorageInstance) SetPicture(device *storage.Device, picture *Picture) {
	instance.writeRequestChannel <- writeRequest{
		device:  device,
		picture: picture,
	}
}

func (instance *PictureStorageInstance) GetPicture(device *storage.Device) (picture *Picture, err error) {
	response := make(chan error)
	instance.readRequestChannel <- readRequest{
		device:   device,
		picture:  picture,
		response: response,
	}
	err = <-response
	return
}
