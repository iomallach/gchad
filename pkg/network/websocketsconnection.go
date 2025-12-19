package network

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/pkg/logging"
)

type WebsocketsConnection struct {
	conn   *websocket.Conn
	logger logging.Logger
}

func NewWebsocketsConnection(conn *websocket.Conn, logger logging.Logger) *WebsocketsConnection {
	return &WebsocketsConnection{conn, logger}
}

func (ws *WebsocketsConnection) Close() error {
	return ws.conn.Close()
}

func (ws *WebsocketsConnection) ReadMessage() (int, []byte, error) {
	bytesRead, msg, err := ws.conn.ReadMessage()
	if err != nil {
		return bytesRead, msg, TranslateReadError(err)
	}

	ws.logger.Debug("read message", map[string]any{"bytes_read": bytesRead, "message": string(msg)})

	return bytesRead, msg, nil
}

func (ws *WebsocketsConnection) SetWriteDeadline(t time.Time) error {
	return ws.conn.SetWriteDeadline(t)
}

func (ws *WebsocketsConnection) SetReadDeadline(t time.Time) error {
	return ws.conn.SetReadDeadline(t)
}

func (ws *WebsocketsConnection) SetPongHandler(f func(string) error) {
	ws.conn.SetPongHandler(f)
}

func (ws *WebsocketsConnection) SetPingHandler(f func(string) error) {
	ws.conn.SetPingHandler(f)
}

func (ws *WebsocketsConnection) writeMessage(msgCode int, data []byte) error {
	err := ws.conn.WriteMessage(msgCode, data)
	return TranslateWriteError(err)
}

func (ws *WebsocketsConnection) WriteCloseMessage(data []byte) error {
	return ws.writeMessage(websocket.CloseMessage, data)
}

func (ws *WebsocketsConnection) WriteTextMessage(data []byte) error {
	return ws.writeMessage(websocket.TextMessage, data)
}

func (ws *WebsocketsConnection) WritePingMessage(data []byte) error {
	return ws.writeMessage(websocket.PingMessage, data)
}

func (ws *WebsocketsConnection) WritePongMessage(data []byte) error {
	return ws.writeMessage(websocket.PongMessage, data)
}
