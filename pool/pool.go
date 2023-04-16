package pool

import (
	"sync"
)

type Poolable interface {
	Name() string
	Shutdown()
}

type Pool[I Poolable] struct {
	items map[string]I
	mutex sync.RWMutex
}

func RunPool[I Poolable]() *Pool[I] {
	return &Pool[I]{
		items: make(map[string]I),
	}
}

func (p *Pool[I]) Shutdown() {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	for _, c := range p.items {
		c.Shutdown()
	}
}

func (p *Pool[I]) AddDevice(item I) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.items[item.Name()] = item
}

func (p *Pool[I]) RemoveDevice(item I) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.items, item.Name())
}

func (p *Pool[I]) GetDevice(deviceName string) I {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.items[deviceName]
}

func (p *Pool[I]) GetDevices() map[string]I {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	ret := make(map[string]I, len(p.items))
	for name, item := range p.items {
		ret[name] = item
	}
	return ret
}
