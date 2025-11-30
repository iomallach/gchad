package application

import "github.com/iomallach/gchad/internal/domain"

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

func (r *ClientRegistry) RemoveClient(client *domain.Client) {
	delete(r.clients, client.Id())
}

func (r *ClientRegistry) GetAllClients() []*domain.Client {
	clients := make([]*domain.Client, len(r.clients))

	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}
