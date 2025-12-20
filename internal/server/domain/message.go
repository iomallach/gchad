package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType string

const (
	SystemUserJoined   MessageType = "user_joined"
	SystemUserLeft     MessageType = "user_left"
	UserMsg            MessageType = "chat"
	SystemStatsMessage MessageType = "stats"
)

type Messager interface {
	MessageType() MessageType
}

type StatsSystemMessage struct {
	ClientsOnline int `json:"clients_online"`
}

// TODO: session duration not implemented yet
func NewStatsSystemMessage(clientsOnline int) *StatsSystemMessage {
	return &StatsSystemMessage{
		ClientsOnline: clientsOnline,
	}
}

func (m *StatsSystemMessage) MessageType() MessageType {
	return SystemStatsMessage
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
	Text      string    `json:"text"`
	From      string    `json:"from"`
}

func NewUserMessage(msg string, timestamp time.Time, from string) *UserMessage {
	return &UserMessage{
		Timestamp: timestamp,
		Text:      msg,
		From:      from,
	}
}

func (m *UserMessage) MessageType() MessageType {
	return UserMsg
}

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func UnmarshalMessage(data []byte) (Messager, error) {
	var envelope Message

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	var msg Messager
	switch envelope.Type {
	case SystemUserJoined:
		msg = &UserJoinedSystemMessage{}
	case SystemUserLeft:
		msg = &UserLeftSystemMessage{}
	case UserMsg:
		msg = &UserMessage{}
	case SystemStatsMessage:
		msg = &StatsSystemMessage{}
	default:
		return nil, fmt.Errorf("unknown message %s of type: %s", envelope.Payload, envelope.Type)
	}

	if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
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
		Type:    data.MessageType(),
		Payload: json_data,
	}

	return json.Marshal(envelope)
}
