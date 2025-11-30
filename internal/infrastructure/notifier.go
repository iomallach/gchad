package infrastructure

import (
	"sync"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type Notifier interface {
	BroadcastToRoom(*application.ChatRoom, domain.Messager) error
	RegisterClient(*ClientAdapter)
	UnregisterClient(string)
}

type ClientNotifier struct {
	mu      sync.RWMutex
	clients map[string]*ClientAdapter
}

func NewClientNotifier() *ClientNotifier {
	return &ClientNotifier{
		mu:      sync.RWMutex{},
		clients: make(map[string]*ClientAdapter),
	}
}

func (n *ClientNotifier) BroadcastToRoom(room *application.ChatRoom, msg domain.Messager) error {
	message, err := domain.MarshallMessage(msg)
	if err != nil {
		// TODO: log error
		return err
	}
	n.mu.RLock()
	defer n.mu.Unlock()

	for _, client := range room.GetClients() {
		if adapter, ok := n.clients[client.Id()]; ok {
			select {
			// TODO: need to send the actual message here
			case adapter.send <- message:
			default:
			}
		} else {
			// TODO: need to log this
		}
	}

	return nil
}

func (n *ClientNotifier) RegisterClient(adapter *ClientAdapter) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.clients[adapter.Id()] = adapter
}

func (n *ClientNotifier) UnregisterClient(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.clients, id)
}
