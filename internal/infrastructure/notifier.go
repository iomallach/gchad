package infrastructure

import (
	"sync"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type ClientNotifier struct {
	mu         sync.RWMutex
	clients    map[string]ClientAdapter
	register   chan ClientAdapter
	unregister chan string
	logger     application.Logger
}

func NewClientNotifier(logger application.Logger) *ClientNotifier {
	return &ClientNotifier{
		mu:         sync.RWMutex{},
		clients:    make(map[string]ClientAdapter),
		register:   make(chan ClientAdapter, 256),
		unregister: make(chan string, 256),
		logger:     logger,
	}
}

func (n *ClientNotifier) Start() {
	go n.handleClientLifecycle()
}

func (n *ClientNotifier) BroadcastToRoom(room *application.ChatRoom, msg domain.Messager) error {
	n.mu.RLock()
	defer n.mu.Unlock()

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

	return nil
}

func (n *ClientNotifier) RegisterClient(client ClientAdapter) {
	select {
	case n.register <- client:
		n.logger.Debug("scheduled client to be registered", map[string]any{"client_id": client.Id()})
	default:
		n.logger.Error("channel is full or closed", map[string]any{"client_id": client.Id()})
	}
}

func (n *ClientNotifier) UnregisterClient(id string) {
	select {
	case n.unregister <- id:
		n.logger.Debug("scheduled client to be unregistered", map[string]any{"client_id": id})
	default:
		n.logger.Error("channel is full or closed", map[string]any{"client_id": id})
	}
}

func (n *ClientNotifier) handleClientLifecycle() {
	for {
		select {
		case client := <-n.register:
			// TODO: handle already exists situation
			n.mu.Lock()
			n.clients[client.Id()] = client
			n.mu.Unlock()
		case clientId := <-n.unregister:
			// TODO: handle doesn't exist situation
			n.mu.Lock()
			delete(n.clients, clientId)
			n.mu.Unlock()
		}
	}
}
