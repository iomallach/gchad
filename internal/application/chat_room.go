package application

import (
	"sync"

	"github.com/iomallach/gchad/internal/domain"
)

type ChatRoom struct {
	mu      sync.RWMutex
	id      string
	name    string
	clients *ClientRegistry
}

func NewChatRoom(id string, name string, clients *ClientRegistry) *ChatRoom {
	return &ChatRoom{
		mu:      sync.RWMutex{},
		id:      id,
		name:    name,
		clients: clients,
	}
}

func (cr *ChatRoom) Id() string {
	return cr.id
}

func (cr *ChatRoom) Name() string {
	return cr.name
}

func (cr *ChatRoom) LetClientIn(client *domain.Client) *domain.UserJoinedRoom {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.clients.AddClient(client)

	return domain.NewUserJoinedRoomEvent(client.Name())
}

func (cr *ChatRoom) LetClientOut(clientId string) *domain.UserLeftRoom {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// TODO: null pointer exception danger
	client := cr.clients.GetClient(clientId)
	cr.clients.RemoveClient(clientId)

	return domain.NewUserLeftRoomEvent(client.Name())
}

func (cr *ChatRoom) GetClients() []*domain.Client {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	return cr.clients.GetAllClients()
}
