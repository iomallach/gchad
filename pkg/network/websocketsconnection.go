package network

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/server/application"
)

var (
	ErrConnectionClosedNormally   = errors.New("connection closed normally")
	ErrConnectionClosedAbnormally = errors.New("connection closed abnormally")
	ErrMessageTooLarge            = errors.New("message exceeds size limit")
	ErrNetworkFailure             = errors.New("network failure")
	ErrReadTimeOut                = errors.New("read timeout")

	ErrWriteAfterClose = errors.New("write after close")
	ErrWriteTimeout    = errors.New("write timeout")
	ErrWriteFailed     = errors.New("write unexpectedly failed")
)

type WebsocketsConnection struct {
	conn   *websocket.Conn
	logger application.Logger
}

func NewWebsocketsConnection(conn *websocket.Conn, logger application.Logger) *WebsocketsConnection {
	return &WebsocketsConnection{conn, logger}
}

func (ws *WebsocketsConnection) Close() error {
	return ws.conn.Close()
}

func (ws *WebsocketsConnection) ReadMessage() (int, []byte, error) {
	bytesRead, msg, err := ws.conn.ReadMessage()
	if err != nil {
		return bytesRead, msg, translateReadError(err)
	}

	ws.logger.Debug("read message", map[string]any{"bytes_read": bytesRead, "message": string(msg)})

	return bytesRead, msg, nil
}

func (ws *WebsocketsConnection) SetWriteDeadline(t time.Time) error {
	return ws.conn.SetWriteDeadline(t)
}

func (ws *WebsocketsConnection) writeMessage(msgCode int, data []byte) error {
	err := ws.conn.WriteMessage(msgCode, data)
	return translateWriteError(err)
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

func translateWriteError(err error) error {
	if err == nil {
		return nil
	}

	if err == websocket.ErrCloseSent {
		return ErrWriteAfterClose
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrWriteTimeout
	}

	return ErrWriteFailed
}

func translateReadError(err error) error {
	if websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
	) {
		return ErrConnectionClosedNormally
	}

	if websocket.IsCloseError(
		err,
		websocket.CloseProtocolError,
		websocket.CloseUnsupportedData,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
		websocket.CloseInvalidFramePayloadData,
		websocket.CloseInternalServerErr,
		websocket.CloseTryAgainLater,
	) {
		return ErrConnectionClosedAbnormally
	}

	if _, ok := err.(*websocket.CloseError); ok {
		return ErrConnectionClosedAbnormally
	}

	if err == io.EOF {
		return ErrConnectionClosedAbnormally
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrReadTimeOut
	}

	if err == websocket.ErrReadLimit {
		return ErrMessageTooLarge
	}

	return ErrNetworkFailure
}
