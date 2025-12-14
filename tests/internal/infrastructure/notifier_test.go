package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/iomallach/gchad/internal/server/application"
	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/iomallach/gchad/internal/server/infrastructure"
	"github.com/stretchr/testify/assert"
)

func TestClientNotifier_BroadCastToRoom(t *testing.T) {
	clientConfiguration := infrastructure.ClientConfiguration{
		SendChannelSize: 1,
	}
	clients := []*domain.Client{domain.NewClient("1", "Jane Doe"), domain.NewClient("2", "John Doe")}
	adapters := make([]*infrastructure.Client, 0)
	adapterSpyLogger := SpyLogger{calls: make([]LogCall, 0)}
	for _, client := range clients {
		adapters = append(
			adapters,
			infrastructure.NewClient(
				client.Id(),
				client.Name(),
				nil,
				nil,
				make(chan domain.Messager, clientConfiguration.SendChannelSize),
				clientConfiguration,
				&adapterSpyLogger,
			),
		)
	}
	spyLogger := SpyLogger{calls: make([]LogCall, 0)}
	clientRegistry := application.NewClientRegistry()
	room := application.NewChatRoom("1", "general", clientRegistry)
	for _, client := range clients {
		room.LetClientIn(client)
	}
	existingClients := make(map[string]*infrastructure.Client, 0)
	for _, adapter := range adapters {
		existingClients[adapter.Id()] = adapter
	}
	notifier := infrastructure.NewClientNotifierFromExistingClients(&spyLogger, existingClients)

	notifier.BroadcastToRoom(room, domain.NewUserMessage("Hello test", time.Now(), "test"))

	for _, adapter := range adapters {
		msgIntf := <-adapter.Send()
		msg, ok := msgIntf.(*domain.UserMessage)
		if !ok {
			t.Errorf("expected a user message, got %T", msgIntf)
		}

		assert.Equal(t, "Hello test", msg.Message)
	}
}

func TestClientNotifier_RegisterUnregisterClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientConfiguration := infrastructure.ClientConfiguration{
		SendChannelSize: 1,
	}
	adapterSpyLogger := SpyLogger{calls: make([]LogCall, 0)}
	clients := []*infrastructure.Client{
		infrastructure.NewClient("1", "Jane Doe", nil, nil, nil, clientConfiguration, &adapterSpyLogger),
		infrastructure.NewClient("2", "John Doe", nil, nil, nil, clientConfiguration, &adapterSpyLogger),
	}
	spyLogger := SpyLogger{calls: make([]LogCall, 0)}
	registry := make(map[string]*infrastructure.Client)
	notifier := infrastructure.NewClientNotifier(&spyLogger, registry)

	notifier.Start(ctx)

	// first register all the clients
	for _, client := range clients {
		notifier.RegisterClient(client)
	}

	// TODO: can I replaces these inefficient occurences with a select?
	time.Sleep(time.Millisecond * 50)

	assert.Len(t, spyLogger.Errors(), 0)
	assert.Len(t, spyLogger.Debugs(), 2)
	assert.Len(t, registry, 2)
	for _, client := range clients {
		clientFromRegistry, ok := registry[client.Id()]
		if !ok {
			t.Fatalf("expected client %s to be in registry, but it wasn't there", client.Id())
		}
		assert.Equal(t, client, clientFromRegistry)
	}

	// then unregister
	for _, client := range clients {
		notifier.UnregisterClient(client.Id())
	}

	time.Sleep(time.Millisecond * 50)

	assert.Len(t, spyLogger.Errors(), 0)
	assert.Len(t, registry, 0)
}
