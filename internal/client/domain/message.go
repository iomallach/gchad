package domain

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	TypeChatMessage       MessageType = "chat"
	TypeUserJoinedMessage MessageType = "user_joined"
	TypeUserLeftMessage   MessageType = "user_left"
)

type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ChatMessage struct {
	From      string    `json:"from"`
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
}

func (m ChatMessage) MessageType() MessageType {
	return TypeChatMessage
}

type UserJoinedMessage struct {
	Timestamp time.Time
	Name      string
}

func (m UserJoinedMessage) MessageType() MessageType {
	return TypeUserJoinedMessage
}

type UserLeftMessage struct {
	Timestamp time.Time
	Name      string
}

func (m UserLeftMessage) MessageType() MessageType {
	return TypeUserLeftMessage
}
