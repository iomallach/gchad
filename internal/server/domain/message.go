package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType string

const (
	SystemUserJoined MessageType = "user_joined"
	SystemUserLeft   MessageType = "user_left"
	UserMsg          MessageType = "user_message"
	NewConnection    MessageType = "connect_me"
)

type Messager interface {
	MessageType() MessageType
}

type UserJoinedSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func NewUserJoinedSystemMessage(name string, timestamp time.Time) *UserJoinedSystemMessage {
	return &UserJoinedSystemMessage{
		Timestamp: timestamp,
		Name:      name,
	}
}

func (m *UserJoinedSystemMessage) MessageType() MessageType {
	return SystemUserJoined
}

type UserLeftSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func NewUserLeftSystemMessage(name string, timestamp time.Time) *UserLeftSystemMessage {
	return &UserLeftSystemMessage{
		Timestamp: timestamp,
		Name:      name,
	}
}

func (m *UserLeftSystemMessage) MessageType() MessageType {
	return SystemUserLeft
}

type UserMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	From      string    `json:"from"`
}

func NewUserMessage(msg string, timestamp time.Time, from string) *UserMessage {
	return &UserMessage{
		Timestamp: timestamp,
		Message:   msg,
		From:      from,
	}
}

func (m *UserMessage) MessageType() MessageType {
	return UserMsg
}

type Message struct {
	MessageType MessageType     `json:"type"`
	Data        json.RawMessage `json:"data"`
}

func UnmarshalMessage(data []byte) (Messager, error) {
	var envelope Message

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	var msg Messager
	switch envelope.MessageType {
	case SystemUserJoined:
		msg = &UserJoinedSystemMessage{}
	case SystemUserLeft:
		msg = &UserLeftSystemMessage{}
	case UserMsg:
		msg = &UserMessage{}
	default:
		return nil, fmt.Errorf("unknown message type: %s", envelope.MessageType)
	}

	if err := json.Unmarshal(envelope.Data, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func MarshallMessage(data Messager) ([]byte, error) {
	json_data, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	envelope := Message{
		MessageType: data.MessageType(),
		Data:        json_data,
	}

	return json.Marshal(envelope)
}
