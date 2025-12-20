package application

import "github.com/iomallach/gchad/internal/client/domain"

type Message interface {
	MessageType() domain.MessageType
}

type ChatClient interface {
	Connect() error
	Disconnect() error
	SendMessage(message string)
	InboundMessages() <-chan Message
	Errors() <-chan error
	SetName(name string)
	Host() string
	Stats() *domain.ChatStats
}
