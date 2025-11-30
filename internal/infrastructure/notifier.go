package infrastructure

import (
	"sync"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type Notifier interface {
	application.Notifier
	RegisterClient(*ClientAdapter)
	UnregisterClient(string)
}

type ClientNotifier struct {
	mu      sync.RWMutex
	clients map[string]*ClientAdapter
	logger  Logger
}

func NewClientNotifier(logger Logger) *ClientNotifier {
	return &ClientNotifier{
		mu:      sync.RWMutex{},
		clients: make(map[string]*ClientAdapter),
		logger:  logger,
	}
}

func (n *ClientNotifier) BroadcastToRoom(room *application.ChatRoom, msg domain.Messager) error {
	message, err := domain.MarshallMessage(msg)
	if err != nil {
		n.logger.Error().Err(err).Msg("failed to serialize a message")
		return err
	}
	n.mu.RLock()
	defer n.mu.Unlock()

	for _, client := range room.GetClients() {
		if adapter, ok := n.clients[client.Id()]; ok {
			select {
			case adapter.send <- message:
				n.logger.Debug().Str("client_id", client.Id()).Str("message_type", string(msg.MessageType())).Msg("message queued for client")
			default:
				n.logger.Error().Str("client_id", client.Id()).Msg("failed to queue message, channel is full or closed")
			}
		} else {
			n.logger.Debug().Str("client_id", client.Id()).Msg("attempted to broadcast to client that doesn't exist")
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
