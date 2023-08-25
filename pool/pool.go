package pool

import (
	"golang.org/x/exp/maps"
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

func (p *Pool[I]) Add(item I) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.items[item.Name()] = item
}

func (p *Pool[I]) Remove(item I) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.items, item.Name())
}

func (p *Pool[I]) GetAll() map[string]I {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return maps.Clone(p.items)
}

func (p *Pool[I]) GetByName(name string) I {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.items[name]
}

func (p *Pool[I]) GetByNames(names []string) []I {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// copy only those in the names list
	ret := make([]I, 0, len(p.items))
	for _, name := range names {
		if i, ok := p.items[name]; ok {
			ret = append(ret, i)
		}
	}
	return ret
}
