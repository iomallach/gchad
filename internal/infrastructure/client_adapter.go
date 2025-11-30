package infrastructure

import "time"

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
	id   string
	name string
	conn Connection
	// TODO: ChatService?
	notifier      Notifier
	send          chan []byte
	configuration ClientConfiguration
}

func (c *ClientAdapter) Id() string {
	return c.id
}

func (c *ClientAdapter) Send() chan []byte {
	return c.send
}

func NewClientAdapter(id string, name string, conn Connection, notifier Notifier, configuration ClientConfiguration) *ClientAdapter {
	return &ClientAdapter{
		id:            id,
		name:          name,
		conn:          conn,
		notifier:      notifier,
		send:          make(chan []byte, configuration.sendChannelSize),
		configuration: configuration,
	}
}
