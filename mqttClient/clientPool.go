package mqttClient

import (
	"sync"
)

type ClientPool struct {
	Clients      map[string]*Client
	ClientsMutex sync.RWMutex
}

func RunPool() (pool *ClientPool) {
	pool = &ClientPool{
		Clients: make(map[string]*Client),
	}
	return
}

func (p *ClientPool) Shutdown() {
	p.ClientsMutex.RLock()
	defer p.ClientsMutex.RUnlock()
	for _, c := range p.Clients {
		c.Shutdown()
	}
}

func (p *ClientPool) AddClient(client *Client) {
	p.ClientsMutex.Lock()
	defer p.ClientsMutex.Unlock()
	p.Clients[client.Config().Name()] = client
}

func (p *ClientPool) RemoveClient(client Client) {
	p.ClientsMutex.Lock()
	defer p.ClientsMutex.Unlock()
	delete(p.Clients, client.Config().Name())
}

func (p *ClientPool) GetClient(clientName string) *Client {
	p.ClientsMutex.RLock()
	defer p.ClientsMutex.RUnlock()
	return p.Clients[clientName]
}
