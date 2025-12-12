package infrastructure

import (
	"fmt"
	"time"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
)

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	SetWriteDeadline(time.Time) error
	WriteCloseMessage([]byte) error
	WriteTextMessage([]byte) error
	WritePingMessage([]byte) error
}

type ClientConfiguration struct {
	WriteWait          time.Duration
	PongWait           time.Duration
	PingPeriod         time.Duration
	RecieveChanWait    time.Duration
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
	logger        application.Logger
}

func (c *ClientAdapter) Id() string {
	return c.id
}

func (c *ClientAdapter) Send() chan domain.Messager {
	return c.send
}

func (c *ClientAdapter) Start() {
	go c.ReadMessages()
	go c.WriteMessages()
}

func NewClientAdapter(id string, name string, conn Connection, configuration ClientConfiguration, logger application.Logger) *ClientAdapter {
	return &ClientAdapter{
		id:            id,
		name:          name,
		conn:          conn,
		send:          make(chan domain.Messager, configuration.SendChannelSize),
		recv:          make(chan domain.Messager, configuration.ReceiveChannelSize),
		configuration: configuration,
		logger:        logger,
	}
}

// TODO: Need to figure out graceful shutdown of both pumps
func (c *ClientAdapter) ReadMessages() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("could not read the message", map[string]any{"client_id": c.Id()})
			continue
		}

		domainMessage, err := domain.UnmarshalMessage(message)
		if err != nil {
			c.logger.Error("failed to unmarshall the message", map[string]any{"client_id": c.Id()})
			continue
		}

		select {
		case c.recv <- domainMessage:
			c.logger.Debug("message received", make(map[string]any))
		case <-time.After(c.configuration.RecieveChanWait):
			c.logger.Error("message channel is full, skipping message", map[string]any{"client_id": c.Id()})
		}
	}
}

func (c *ClientAdapter) WriteMessages() {
	ticker := time.NewTicker(c.configuration.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case domanMessage, ok := <-c.send:
			if !ok {
				c.logger.Error(
					"send channel has been closed. Sending close message and terminating",
					map[string]any{"client_id": c.Id()},
				)
				err := c.conn.WriteCloseMessage([]byte{})
				if err != nil {
					c.logger.Error(fmt.Sprintf("failed to write close message: %s", err.Error()), map[string]any{"client_id": c.Id()})
				}
				return
			}

			c.logger.Debug("sending message", map[string]any{"client_id": c.Id()})

			message, err := domain.MarshallMessage(domanMessage)
			if err != nil {
				c.logger.Error("failed to marshall a message", map[string]any{"client_id": c.Id()})
			}

			err = c.conn.WriteTextMessage(message)
			if err != nil {
				c.logger.Error(fmt.Sprintf("failed to write message: %s", err.Error()), map[string]any{"client_id": c.Id()})
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.configuration.WriteWait))
			if err := c.conn.WritePingMessage(nil); err != nil {
				c.logger.Error(fmt.Sprintf("failed to write ping message: %s", err.Error()), map[string]any{"client_id": c.Id()})
				return
			}
		}
	}
}
