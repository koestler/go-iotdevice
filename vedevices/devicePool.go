package vedevices

import "sync"

type DevicePool struct {
	Devices      map[string]*Device
	DevicesMutex sync.RWMutex
}

func RunPool() (pool *DevicePool) {
	pool = &DevicePool{
		Devices: make(map[string]*Device),
	}
	return
}

func (p *DevicePool) Shutdown() {
	p.DevicesMutex.RLock()
	defer p.DevicesMutex.RUnlock()
	for _, c := range p.Devices {
		c.Shutdown()
	}
}

func (p *DevicePool) AddDevice(device *Device) {
	p.DevicesMutex.Lock()
	defer p.DevicesMutex.Unlock()
	p.Devices[device.Name()] = device
}

func (p *DevicePool) RemoveDevice(device *Device) {
	p.DevicesMutex.Lock()
	defer p.DevicesMutex.Unlock()
	delete(p.Devices, device.Name())
}

func (p *DevicePool) GetDevice(deviceName string) *Device {
	p.DevicesMutex.RLock()
	defer p.DevicesMutex.RUnlock()
	return p.Devices[deviceName]
}
