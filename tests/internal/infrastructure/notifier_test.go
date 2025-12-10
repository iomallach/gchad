package infrastructure_test

import (
	"testing"
	"time"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
	"github.com/iomallach/gchad/internal/infrastructure"
	"github.com/stretchr/testify/assert"
)

type LogCall struct {
	msg    string
	fields map[string]any
	level  string
}

type SpyLogger struct {
	calls []LogCall
}

func (l *SpyLogger) Error(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "ERROR"})
}
func (l *SpyLogger) Info(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "INFO"})
}
func (l *SpyLogger) Debug(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "DEBUG"})
}

func (l *SpyLogger) Errors() []LogCall {
	errors := make([]LogCall, 0)

	for _, call := range l.calls {
		if call.level == "ERROR" {
			errors = append(errors, call)
		}
	}

	return errors
}

func (l *SpyLogger) Debugs() []LogCall {
	errors := make([]LogCall, 0)

	for _, call := range l.calls {
		if call.level == "DEBUG" {
			errors = append(errors, call)
		}
	}

	return errors
}

func TestClientNotifier_BroadCastToRoom(t *testing.T) {
	clientConfiguration := infrastructure.ClientConfiguration{
		SendChannelSize:    1,
		ReceiveChannelSize: 1,
	}
	clients := []*domain.Client{domain.NewClient("1", "Jane Doe"), domain.NewClient("2", "John Doe")}
	adapters := make([]*infrastructure.ClientAdapter, 0)
	for _, client := range clients {
		adapters = append(adapters, infrastructure.NewClientAdapter(client.Id(), client.Name(), nil, clientConfiguration))
	}
	spyLogger := SpyLogger{calls: make([]LogCall, 0)}
	clientRegistry := application.NewClientRegistry()
	room := application.NewChatRoom("1", "general", clientRegistry)
	for _, client := range clients {
		room.LetClientIn(client)
	}
	existingClients := make(map[string]*infrastructure.ClientAdapter, 0)
	for _, adapter := range adapters {
		existingClients[adapter.Id()] = adapter
	}
	notifier := infrastructure.NewClientNotifierFromExistingClients(&spyLogger, existingClients)

	notifier.BroadcastToRoom(room, domain.NewUserMessage("Hello test", time.Now()))

	for _, adapter := range adapters {
		msgIntf := <-adapter.Send()
		msg, ok := msgIntf.(*domain.UserMessage)
		if !ok {
			t.Errorf("expected a user message, got %T", msgIntf)
		}

		assert.Equal(t, "Hello test", msg.Message)
	}
}
