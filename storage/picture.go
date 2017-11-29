package storage

import (
	"errors"
	"time"
)

var PictureDb = PictureStorageCreate()

type Picture struct {
	Created time.Time
	Path    string
}

type State map[*Device]*Picture

type writeRequest struct {
	device  *Device
	picture *Picture
}

type readResponse struct {
	picture *Picture
	err     error
}

type readRequest struct {
	device   *Device
	response chan readResponse
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
				readRequest.response <- readResponse{
					picture: picture,
					err:     nil,
				}
			} else {
				readRequest.response <- readResponse{
					picture: nil,
					err:     errors.New("picture for given device does not exist"),
				}
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

func (instance *PictureStorageInstance) SetPicture(device *Device, picture *Picture) {
	instance.writeRequestChannel <- writeRequest{
		device:  device,
		picture: picture,
	}
}

func (instance *PictureStorageInstance) GetPicture(device *Device) (picture *Picture, err error) {
	response := make(chan readResponse)
	instance.readRequestChannel <- readRequest{
		device:   device,
		response: response,
	}
	r := <-response
	return r.picture, r.err
}
