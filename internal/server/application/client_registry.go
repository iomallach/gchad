package application

import (
	"sync"

	"github.com/iomallach/gchad/internal/server/domain"
)

type ClientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*domain.Client
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients: make(map[string]*domain.Client),
		mu:      sync.RWMutex{},
	}
}

func (r *ClientRegistry) AddClient(client *domain.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.Id()] = client
}

func (r *ClientRegistry) RemoveClient(clientId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, clientId)
}

func (r *ClientRegistry) GetAllClients() []*domain.Client {
	r.mu.Lock()
	defer r.mu.Unlock()

	clients := make([]*domain.Client, 0, len(r.clients))

	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *ClientRegistry) GetClient(clientId string) *domain.Client {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.clients[clientId]
}
