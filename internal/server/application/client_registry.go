package application

import "github.com/iomallach/gchad/internal/server/domain"

type ClientRegistry struct {
	clients map[string]*domain.Client
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients: make(map[string]*domain.Client),
	}
}

func (r *ClientRegistry) AddClient(client *domain.Client) {
	r.clients[client.Id()] = client
}

func (r *ClientRegistry) RemoveClient(clientId string) {
	delete(r.clients, clientId)
}

func (r *ClientRegistry) GetAllClients() []*domain.Client {
	clients := make([]*domain.Client, 0, len(r.clients))

	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *ClientRegistry) GetClient(clientId string) *domain.Client {
	return r.clients[clientId]
}
