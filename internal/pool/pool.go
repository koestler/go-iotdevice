package pool

import "sync"

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

func (p *Pool[I]) Add(item I) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.items[item.Name()] = item
}

func (p *Pool[I]) GetByName(name string) (item I, ok bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	item, ok = p.items[name]
	return
}
