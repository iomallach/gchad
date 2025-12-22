package application

import (
	"github.com/iomallach/gchad/internal/server/domain"
)

type ChatRoom struct {
	id      string
	name    string
	clients *ClientRegistry
}

func NewChatRoom(id string, name string, clients *ClientRegistry) *ChatRoom {
	return &ChatRoom{
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
	cr.clients.AddClient(client)

	return domain.NewUserJoinedRoomEvent(client.Name())
}

func (cr *ChatRoom) LetClientOut(clientId string) *domain.UserLeftRoom {
	// TODO: null pointer exception danger
	client := cr.clients.GetClient(clientId)
	cr.clients.RemoveClient(clientId)

	return domain.NewUserLeftRoomEvent(client.Name())
}

func (cr *ChatRoom) GetClients() []*domain.Client {
	return cr.clients.GetAllClients()
}
