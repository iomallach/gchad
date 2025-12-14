package application

import "github.com/iomallach/gchad/internal/server/domain"

type Notifier interface {
	BroadcastToRoom(*ChatRoom, domain.Messager)
}
