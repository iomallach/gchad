package gchad

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshallMessage(t *testing.T) {
	timestamp := time.Time{}
	testCases := []struct {
		expected string
		input    Message
	}{
		{
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","message":"Hello world"},"message_type":3}`,
			Message{
				Inner: &UserMessage{
					Timestamp: timestamp,
					Message:   "Hello world",
				},
				MessageType: UserMsg,
			},
		},
		{
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"},"message_type":0}`,
			Message{
				Inner: &UserJoinedSystemMessage{
					Timestamp: timestamp,
					Name:      "John Doe",
				},
				MessageType: SystemUserJoined,
			},
		},
		{
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"},"message_type":1}`,
			Message{
				Inner: &UserLeftSystemMessage{
					Timestamp: timestamp,
					Name:      "John Doe",
				},
				MessageType: SystemUserLeft,
			},
		},
	}

	for _, testCase := range testCases {
		data, err := json.Marshal(testCase.input)
		if err != nil {
			t.Error("Failed to marshal message:", err)
		}

		if string(data) != testCase.expected {
			t.Errorf("Expected marshalled message to be %s, got %s", testCase.expected, data)
		}
	}
}

func TestUnmarshallMessage(t *testing.T) {
	timestamp := time.Time{}
	testCases := []struct {
		expected Message
		input    string
	}{
		{
			Message{
				Inner: &UserMessage{
					Timestamp: timestamp,
					Message:   "Hello world",
				},
				MessageType: UserMsg,
			},
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","message":"Hello world"},"message_type":3}`,
		},
		{
			Message{
				Inner: &UserJoinedSystemMessage{
					Timestamp: timestamp,
					Name:      "John Doe",
				},
				MessageType: SystemUserJoined,
			},
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"},"message_type":0}`,
		},
		{
			Message{
				Inner: &UserLeftSystemMessage{
					Timestamp: timestamp,
					Name:      "John Doe",
				},
				MessageType: SystemUserLeft,
			},
			`{"inner":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"},"message_type":1}`,
		},
	}

	for _, testCase := range testCases {
		var unmarshalled Message
		err := json.Unmarshal([]byte(testCase.input), &unmarshalled)
		if err != nil {
			t.Fatalf("Failed to unmarshal message:")
		}

		switch expected := testCase.expected.Inner.(type) {
		case *UserMessage:
			unmarshalledInner, ok := unmarshalled.Inner.(*UserMessage)
			if !ok {
				t.Fatalf("Expected inner message to be of type *UserMessage, got %T", unmarshalled.Inner)
			}
			if unmarshalledInner.Timestamp != expected.Timestamp {
				t.Errorf("Expected timestamp to be %s, got %s", expected.Timestamp, unmarshalledInner.Timestamp)
			}
			if unmarshalledInner.Message != expected.Message {
				t.Errorf("Expected message to be %s, got %s", expected.Message, unmarshalledInner.Message)
			}
			if unmarshalled.MessageType != testCase.expected.MessageType {
				t.Errorf("Expected message type to be %d, got %d", testCase.expected.MessageType, unmarshalled.MessageType)
			}
		}
	}
}
