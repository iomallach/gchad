package gchad

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType string

const (
	SystemUserJoined MessageType = "user_joined"
	SystemUserLeft   MessageType = "user_left"
	UserMsg                      = "user_message"
)

type Messager interface {
	MessageType() MessageType
}

type UserJoinedSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func (m *UserJoinedSystemMessage) MessageType() MessageType {
	return SystemUserJoined
}

type UserLeftSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func (m *UserLeftSystemMessage) MessageType() MessageType {
	return SystemUserLeft
}

type UserMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
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
