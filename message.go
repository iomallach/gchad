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
	Inner       Messager    `json:"inner"`
	MessageType MessageType `json:"message_type"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	var innerJson json.RawMessage
	var err error

	if m.Inner != nil {
		innerJson, err = json.Marshal(m.Inner)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(
		struct {
			Inner       json.RawMessage `json:"inner"`
			MessageType MessageType     `json:"message_type"`
		}{
			Inner:       innerJson,
			MessageType: m.MessageType,
		},
	)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		Inner       json.RawMessage `json:"inner"`
		MessageType MessageType     `json:"message_type"`
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
