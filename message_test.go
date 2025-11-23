package gchad_test

import (
	"testing"
	"time"

	"github.com/iomallach/gchad"
)

func TestMarshallMessage(t *testing.T) {
	timestamp := time.Time{}
	testCases := []struct {
		name     string
		input    gchad.Messager
		expected string
	}{
		{
			name: "UserMessage",
			input: &gchad.UserMessage{
				Timestamp: timestamp,
				Message:   "Hello world",
			},
			expected: `{"type":"user_message","data":{"timestamp":"0001-01-01T00:00:00Z","message":"Hello world"}}`,
		},
		{
			name: "UserJoinedSystemMessage",
			input: &gchad.UserJoinedSystemMessage{
				Timestamp: timestamp,
				Name:      "John Doe",
			},
			expected: `{"type":"user_joined","data":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"}}`,
		},
		{
			name: "UserLeftSystemMessage",
			input: &gchad.UserLeftSystemMessage{
				Timestamp: timestamp,
				Name:      "John Doe",
			},
			expected: `{"type":"user_left","data":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"}}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := gchad.MarshallMessage(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			if string(data) != testCase.expected {
				t.Errorf("Expected marshalled message to be %s, got %s", testCase.expected, string(data))
			}
		})
	}
}

func TestUnmarshalMessage(t *testing.T) {
	timestamp := time.Time{}
	testCases := []struct {
		name     string
		input    string
		expected gchad.Messager
	}{
		{
			name:  "UserMessage",
			input: `{"type":"user_message","data":{"timestamp":"0001-01-01T00:00:00Z","message":"Hello world"}}`,
			expected: &gchad.UserMessage{
				Timestamp: timestamp,
				Message:   "Hello world",
			},
		},
		{
			name:  "UserJoinedSystemMessage",
			input: `{"type":"user_joined","data":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"}}`,
			expected: &gchad.UserJoinedSystemMessage{
				Timestamp: timestamp,
				Name:      "John Doe",
			},
		},
		{
			name:  "UserLeftSystemMessage",
			input: `{"type":"user_left","data":{"timestamp":"0001-01-01T00:00:00Z","name":"John Doe"}}`,
			expected: &gchad.UserLeftSystemMessage{
				Timestamp: timestamp,
				Name:      "John Doe",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			unmarshalled, err := gchad.UnmarshalMessage([]byte(testCase.input))
			if err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			if unmarshalled.MessageType() != testCase.expected.MessageType() {
				t.Errorf("Expected message type to be %s, got %s", testCase.expected.MessageType(), unmarshalled.MessageType())
			}

			switch expected := testCase.expected.(type) {
			case *gchad.UserMessage:
				unmarshalledInner, ok := unmarshalled.(*gchad.UserMessage)
				if !ok {
					t.Fatalf("Expected message to be of type *UserMessage, got %T", unmarshalled)
				}
				if unmarshalledInner.Timestamp != expected.Timestamp {
					t.Errorf("Expected timestamp to be %v, got %v", expected.Timestamp, unmarshalledInner.Timestamp)
				}
				if unmarshalledInner.Message != expected.Message {
					t.Errorf("Expected message to be %s, got %s", expected.Message, unmarshalledInner.Message)
				}
			case *gchad.UserJoinedSystemMessage:
				unmarshalledInner, ok := unmarshalled.(*gchad.UserJoinedSystemMessage)
				if !ok {
					t.Fatalf("Expected message to be of type *UserJoinedSystemMessage, got %T", unmarshalled)
				}
				if unmarshalledInner.Timestamp != expected.Timestamp {
					t.Errorf("Expected timestamp to be %v, got %v", expected.Timestamp, unmarshalledInner.Timestamp)
				}
				if unmarshalledInner.Name != expected.Name {
					t.Errorf("Expected name to be %s, got %s", expected.Name, unmarshalledInner.Name)
				}
			case *gchad.UserLeftSystemMessage:
				unmarshalledInner, ok := unmarshalled.(*gchad.UserLeftSystemMessage)
				if !ok {
					t.Fatalf("Expected message to be of type *UserLeftSystemMessage, got %T", unmarshalled)
				}
				if unmarshalledInner.Timestamp != expected.Timestamp {
					t.Errorf("Expected timestamp to be %v, got %v", expected.Timestamp, unmarshalledInner.Timestamp)
				}
				if unmarshalledInner.Name != expected.Name {
					t.Errorf("Expected name to be %s, got %s", expected.Name, unmarshalledInner.Name)
				}
			}
		})
	}
}
