package infrastructure

import (
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
