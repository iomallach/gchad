package infrastructure

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/pkg/logging"
	"github.com/iomallach/gchad/pkg/network"
)

type Dialer interface {
	Dial(url string) (network.Connection, error)
}

type WebsocketsDialer struct {
	dialer *websocket.Dialer
	logger logging.Logger
}

func NewWebsocketDialer(dialer *websocket.Dialer, logger logging.Logger) *WebsocketsDialer {
	return &WebsocketsDialer{dialer, logger}
}

func (d *WebsocketsDialer) Dial(url string) (network.Connection, error) {
	conn, _, err := d.dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return network.NewWebsocketsConnection(conn, d.logger), nil
}

type ChatClient struct {
	conn   network.Connection
	dialer Dialer
	logger logging.Logger

	recv   chan any // TODO: use domain stuff here later
	errors chan error

	send chan any
}

func NewChatClient(
	dialer Dialer,
	recv chan any,
	send chan any,
	errors chan error,
	logger logging.Logger,
) *ChatClient {
	return &ChatClient{
		dialer: dialer,
		logger: logger,
		recv:   recv,
		errors: errors,
		send:   send,
	}
}

func (c *ChatClient) Connect(url string) error {
	conn, err := c.dialer.Dial(url)
	if err != nil {
		return err
	}
	c.logger.Info(fmt.Sprintf("Successfully connected to %s", url), map[string]any{})

	c.conn = conn
	return nil
}

func (c *ChatClient) Disconnect() error {
	return c.conn.Close()
}

func (c *ChatClient) SendMessage(message string) {
	select {
	case c.send <- message:
		c.logger.Debug("message sent", map[string]any{"message": message})
	case <-time.After(50 * time.Millisecond):
		c.logger.Error("failed to send message, channel is full", map[string]any{"message": message})
	}
}

func (c *ChatClient) InboundMessages() <-chan any {
	return c.recv
}

func (c *ChatClient) Errors() <-chan error {
	return c.errors
}
