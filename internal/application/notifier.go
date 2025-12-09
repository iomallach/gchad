package application

import "github.com/iomallach/gchad/internal/domain"

type Notifier interface {
	BroadcastToRoom(*ChatRoom, domain.Messager) error
}
