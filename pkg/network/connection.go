package network

import "time"

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetPongHandler(func(string) error)
	WriteCloseMessage([]byte) error
	WriteTextMessage([]byte) error
	WritePingMessage([]byte) error
}
