package infrastructure

import (
	"context"
	"sync"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type ClientNotifier struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	logger     application.Logger
}

func NewClientNotifier(logger application.Logger, registry map[string]*Client) *ClientNotifier {
	return &ClientNotifier{
		mu:         sync.RWMutex{},
		clients:    registry,
		register:   make(chan *Client, 256),
		unregister: make(chan string, 256),
		logger:     logger,
	}
}

func NewClientNotifierFromExistingClients(logger application.Logger, clients map[string]*Client) *ClientNotifier {
	return &ClientNotifier{
		mu:         sync.RWMutex{},
		clients:    clients,
		register:   make(chan *Client, 256),
		unregister: make(chan string, 256),
		logger:     logger,
	}
}

func (n *ClientNotifier) Start(ctx context.Context) {
	go n.handleClientLifecycle(ctx)
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

func (n *ClientNotifier) handleClientLifecycle(ctx context.Context) {
	for {
		select {
		case client := <-n.register:
			n.mu.Lock()
			if _, ok := n.clients[client.Id()]; ok {
				n.logger.Error("client already exists, skipping adding", map[string]any{"client_id": client.Id()})
			} else {
				n.clients[client.Id()] = client
			}
			n.mu.Unlock()
		case clientId := <-n.unregister:
			n.mu.Lock()
			if _, ok := n.clients[clientId]; ok {
				delete(n.clients, clientId)
			} else {
				n.logger.Error("client doesn't exist, skipping unregistering", map[string]any{"client_id": clientId})
			}
			n.mu.Unlock()
		case <-ctx.Done():
			n.logger.Debug("notifier context done, exiting", make(map[string]any, 0))
			return
		}
	}
}
