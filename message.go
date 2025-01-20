package gchad

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType int

const (
	SystemUserJoined MessageType = iota
	SystemUserLeft
	SystemWhoIsInTheRoom
	UserMsg
)

type Messager interface {
	MessageMark()
}

type UserJoinedSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func (u *UserJoinedSystemMessage) MessageMark() {}

type UserLeftSystemMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
}

func (u *UserLeftSystemMessage) MessageMark() {}

type UserMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

func (u *UserMessage) MessageMark() {}

type Message struct {
	Inner       Messager
	MessageType MessageType
}

func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Inner       interface{} `json:"inner"`
			MessageType MessageType `json:"message_type"`
		}{
			Inner:       m.Inner,
			MessageType: m.MessageType,
		},
	)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		Inner       json.RawMessage `json:"Inner"`
		MessageType MessageType     `json:"MessageType"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.MessageType = raw.MessageType

	var inner Messager
	switch raw.MessageType {
	case SystemUserJoined:
		inner = &UserJoinedSystemMessage{}
	case SystemUserLeft:
		inner = &UserLeftSystemMessage{}
	case UserMsg:
		inner = &UserMessage{}
	default:
		return fmt.Errorf("unknown message type: %d", raw.MessageType)
	}

	if err := json.Unmarshal(raw.Inner, inner); err != nil {
		return err
	}

	m.Inner = inner
	return nil
}
