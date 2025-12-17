package network

import "time"

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	SetWriteDeadline(time.Time) error
	WriteCloseMessage([]byte) error
	WriteTextMessage([]byte) error
	WritePingMessage([]byte) error
}
