package infrastructure

import (
	"sync"

	"github.com/iomallach/gchad/internal/server/application"
	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/iomallach/gchad/pkg/logging"
)

type ClientNotifier struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	logger     logging.Logger
}

func NewClientNotifier(logger logging.Logger, registry map[string]*Client) *ClientNotifier {
	return &ClientNotifier{
		mu:         sync.RWMutex{},
		clients:    registry,
		register:   make(chan *Client, 256),
		unregister: make(chan string, 256),
		logger:     logger,
	}
}

func NewClientNotifierFromExistingClients(logger logging.Logger, clients map[string]*Client) *ClientNotifier {
	return &ClientNotifier{
		mu:         sync.RWMutex{},
		clients:    clients,
		register:   make(chan *Client, 256),
		unregister: make(chan string, 256),
		logger:     logger,
	}
}

func (n *ClientNotifier) BroadcastToRoom(room *application.ChatRoom, msg domain.Messager) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, client := range room.GetClients() {
		if adapter, ok := n.clients[client.Id()]; ok {
			select {
			case adapter.send <- msg:
				n.logger.Debug("message queued for client", map[string]any{"client_id": client.Id(), "message_type": string(msg.MessageType())})
			default:
				n.logger.Error("failed to queue message, channel is full or closed", map[string]any{"client_id": client.Id()})
			}
		} else {
			n.logger.Debug("attempted to broadcast to client that doesn't exist", map[string]any{"client_id": client.Id()})
		}
	}
}

func (n *ClientNotifier) RegisterClient(client *Client) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, ok := n.clients[client.Id()]; ok {
		n.logger.Error("client already exists, skipping adding", map[string]any{"client_id": client.Id()})
	} else {
		n.clients[client.Id()] = client
	}
}

func (n *ClientNotifier) UnregisterClient(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, ok := n.clients[id]; ok {
		delete(n.clients, id)
	} else {
		n.logger.Error("client doesn't exist, skipping unregistering", map[string]any{"client_id": id})
	}
}
