package infrastructure

import (
	"time"

	"github.com/iomallach/gchad/internal/domain"
)

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	SetWriteDeadline(time.Time) error
}

type ClientConfiguration struct {
	WriteWait          time.Duration
	PongWait           time.Duration
	PingPeriod         time.Duration
	SendChannelSize    int
	ReceiveChannelSize int
}

type ClientAdapter struct {
	id            string
	name          string
	conn          Connection
	send          chan domain.Messager
	recv          chan domain.Messager
	configuration ClientConfiguration
}

func (c *ClientAdapter) Id() string {
	return c.id
}

func (c *ClientAdapter) Send() chan domain.Messager {
	return c.send
}

func NewClientAdapter(id string, name string, conn Connection, configuration ClientConfiguration) *ClientAdapter {
	return &ClientAdapter{
		id:            id,
		name:          name,
		conn:          conn,
		send:          make(chan domain.Messager, configuration.SendChannelSize),
		recv:          make(chan domain.Messager, configuration.ReceiveChannelSize),
		configuration: configuration,
	}
}
