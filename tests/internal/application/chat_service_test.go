package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
	"github.com/stretchr/testify/assert"
)

type Broadcast struct {
	room *application.ChatRoom
	msg  domain.Messager
}

type SpyNotifier struct {
	broadcasts []Broadcast
}

func (s *SpyNotifier) BroadcastToRoom(room *application.ChatRoom, msg domain.Messager) error {
	s.broadcasts = append(s.broadcasts, Broadcast{room, msg})

	return nil
}

type ErrorNotifier struct {
	broadcasts []Broadcast
}

func (s *ErrorNotifier) BroadCastToRoom(room *application.ChatRoom, msg domain.Messager) error {
	s.broadcasts = append(s.broadcasts, Broadcast{room, msg})

	return nil
}

func TestChatService_EnterRoom_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientRegistry := application.NewClientRegistry()
	room := application.NewChatRoom("1", "general", clientRegistry)
	frozenTime := time.Date(2025, 12, 7, 0, 0, 0, 0, time.UTC)
	expectedBroadcasts := []struct {
		joined *domain.UserJoinedSystemMessage
		room   *application.ChatRoom
	}{
		{
			joined: &domain.UserJoinedSystemMessage{
				Timestamp: frozenTime,
				Name:      "Jane Doe",
			},
			room: room,
		},
		{
			joined: &domain.UserJoinedSystemMessage{
				Timestamp: frozenTime,
				Name:      "John Doe",
			},
			room: room,
		},
	}

	spyNotifier := SpyNotifier{broadcasts: make([]Broadcast, 0)}
	chatService := application.NewChatService(room, &spyNotifier, func() time.Time { return frozenTime }, 3, 3)

	chatService.Start(ctx)

	chatService.EnterRoom("1", "Jane Doe")
	chatService.EnterRoom("1", "John Doe")

	time.Sleep(50 * time.Millisecond)

	for idx, broadcast := range spyNotifier.broadcasts {
		userJoined, ok := broadcast.msg.(*domain.UserJoinedSystemMessage)
		if !ok {
			t.Errorf("expected a user joined message, got %T", broadcast)
		}
		expected := expectedBroadcasts[idx]
		assert.Equal(t, expected.joined, userJoined)
		assert.Equal(t, expected.room, broadcast.room)
	}
}
