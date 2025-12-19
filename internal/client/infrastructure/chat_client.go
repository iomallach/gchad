package infrastructure

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/client/application"
	"github.com/iomallach/gchad/internal/client/domain"
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

type QueryParam struct {
	key   string
	value string
}

func NewQueryParam(key, value string) QueryParam {
	return QueryParam{key, value}
}

func (qp *QueryParam) String() string {
	return fmt.Sprintf("%s=%s", qp.key, qp.value)
}

type Url struct {
	schema     string
	base       string
	path       string
	port       int
	queryParam QueryParam
}

func NewUrl(schema, base, path string, port int, queryParam QueryParam) Url {
	return Url{schema, base, path, port, queryParam}
}

func (url *Url) String() string {
	subUrl := fmt.Sprintf("%s://%s:%d", url.schema, url.base, url.port)

	if url.path != "" {
		subUrl = fmt.Sprintf("%s/%s", subUrl, url.path)
	}

	return fmt.Sprintf("%s?%s", subUrl, url.queryParam.String())
}

type ChatClient struct {
	conn   network.Connection
	dialer Dialer
	logger logging.Logger

	recv   chan application.Message
	errors chan error

	send chan domain.ChatMessage

	name string
	url  Url
}

func NewChatClient(
	dialer Dialer,
	recv chan application.Message,
	send chan domain.ChatMessage,
	errors chan error,
	logger logging.Logger,
	url Url,
) *ChatClient {
	return &ChatClient{
		dialer: dialer,
		logger: logger,
		recv:   recv,
		errors: errors,
		send:   send,
		url:    url,
	}
}

func (c *ChatClient) Connect() error {
	conn, err := c.dialer.Dial(c.url.String())
	if err != nil {
		return err
	}
	c.logger.Info(fmt.Sprintf("Successfully connected to %s", c.url.String()), map[string]any{})

	c.conn = conn
	go c.ReadPump()
	go c.WritePump()

	return nil
}

func (c *ChatClient) Disconnect() error {
	return c.conn.Close()
}

func (c *ChatClient) SendMessage(message string) {
	select {
	case c.send <- domain.ChatMessage{
		From:      c.name,
		Timestamp: time.Now(),
		Text:      message,
	}:
		c.logger.Debug("message sent", map[string]any{"message": message})
	case <-time.After(50 * time.Millisecond):
		c.logger.Error("failed to send message, channel is full", map[string]any{"message": message})
	}
}

func (c *ChatClient) InboundMessages() <-chan application.Message {
	return c.recv
}

func (c *ChatClient) Errors() <-chan error {
	return c.errors
}

func (c *ChatClient) SetName(name string) {
	c.name = name
	c.url.queryParam.value = name
}

func (c *ChatClient) ReadPump() {
	defer c.conn.Close()

	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(time.Second * 90)); err != nil {
			c.logger.Error(fmt.Sprintf("failed to set read deadline: %s", err.Error()), map[string]any{})
			return
		}

		_, msg, err := c.conn.ReadMessage() // the first value is the ws internal code
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to read a message: %s", err.Error()), map[string]any{})
			return
		}

		message, err := UnmarshallMessage(msg)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to unmarshall the message: %s, %s", msg, err.Error()), map[string]any{})
		}

		select {
		case c.recv <- message:
			c.logger.Debug("new message sent to ui", map[string]any{})
		case <-time.After(time.Millisecond * 50):
			c.logger.Error("message channel is full, skipping sending message to ui", map[string]any{})
		}
	}
}

func (c *ChatClient) WritePump() {
	defer c.conn.Close()

	for msg := range c.send {
		data, err := json.Marshal(msg)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to marshall the message: %s with %s", data, err.Error()), map[string]any{})
			continue
		}

		envelope := domain.Envelope{
			Type:    domain.TypeChatMessage,
			Payload: data,
		}
		message, err := json.Marshal(envelope)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to marshall the message: %s with %s", message, err.Error()), map[string]any{})
			continue
		}

		if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			c.logger.Error(fmt.Sprintf("failed to set write deadline: %s", err.Error()), map[string]any{})
			return
		}
		if err := c.conn.WriteTextMessage(message); err != nil {
			c.logger.Error(fmt.Sprintf("failed to write message: %s", err.Error()), map[string]any{})
			return
		}
		if err := c.conn.SetWriteDeadline(time.Time{}); err != nil {
			c.logger.Error(fmt.Sprintf("failed to clear write deadline: %s", err.Error()), map[string]any{})
			return
		}
	}
}
