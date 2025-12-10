package infrastructure

import (
	"time"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	SetWriteDeadline(time.Time) error
}

type ClientConfiguration struct {
	writeWait       time.Duration
	pongWait        time.Duration
	pingPeriod      time.Duration
	sendChannelSize int
}

type ClientAdapter struct {
	id            string
	name          string
	conn          Connection
	chatService   application.ChatServicer
	send          chan domain.Messager
	configuration ClientConfiguration
}

func (c *ClientAdapter) Id() string {
	return c.id
}

func (c *ClientAdapter) Send() chan domain.Messager {
	return c.send
}

func NewClientAdapter(id string, name string, conn Connection, chatService application.ChatServicer, configuration ClientConfiguration) *ClientAdapter {
	return &ClientAdapter{
		id:            id,
		name:          name,
		conn:          conn,
		chatService:   chatService,
		send:          make(chan domain.Messager, configuration.sendChannelSize),
		configuration: configuration,
	}
}
