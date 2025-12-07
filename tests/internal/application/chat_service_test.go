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

func TestChatService_EnterRoom(t *testing.T) {
	tests := []struct {
		name               string
		clients            []struct{ id, name string }
		expectedBroadcasts func(*application.ChatRoom, time.Time) []Broadcast
	}{
		{
			name: "success with two users joining",
			clients: []struct{ id, name string }{
				{"1", "Jane Doe"},
				{"2", "John Doe"},
			},
			expectedBroadcasts: func(room *application.ChatRoom, frozenTime time.Time) []Broadcast {
				return []Broadcast{
					{
						room: room,
						msg: &domain.UserJoinedSystemMessage{
							Timestamp: frozenTime,
							Name:      "Jane Doe",
						},
					},
					{
						room: room,
						msg: &domain.UserJoinedSystemMessage{
							Timestamp: frozenTime,
							Name:      "John Doe",
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			clientRegistry := application.NewClientRegistry()
			room := application.NewChatRoom("1", "general", clientRegistry)
			frozenTime := time.Date(2025, 12, 7, 0, 0, 0, 0, time.UTC)

			spyNotifier := SpyNotifier{broadcasts: make([]Broadcast, 0)}
			chatService := application.NewChatService(room, &spyNotifier, func() time.Time { return frozenTime }, 3, 3)

			chatService.Start(ctx)

			for _, client := range tt.clients {
				chatService.EnterRoom(client.id, client.name)
			}

			time.Sleep(50 * time.Millisecond)

			expectedBroadcasts := tt.expectedBroadcasts(room, frozenTime)
			assert.Equal(t, len(expectedBroadcasts), len(spyNotifier.broadcasts))

			for idx, broadcast := range spyNotifier.broadcasts {
				userJoined, ok := broadcast.msg.(*domain.UserJoinedSystemMessage)
				if !ok {
					t.Errorf("expected a user joined message, got %T", broadcast)
				}
				assert.Equal(t, expectedBroadcasts[idx].msg, userJoined)
				assert.Equal(t, expectedBroadcasts[idx].room, broadcast.room)
			}
		})
	}
}

func TestChatService_LeaveRoom(t *testing.T) {
	tests := []struct {
		name                 string
		clientsIn            []struct{ id, name string }
		clientsOut           []string
		expectedBroadcasts   func(*application.ChatRoom, time.Time) []Broadcast
		expectedClientsCount int
	}{
		{
			name: "success with two users entering and leaving",
			clientsIn: []struct{ id, name string }{
				{"1", "Jane Doe"},
				{"2", "John Doe"},
			},
			clientsOut: []string{"1", "2"},
			expectedBroadcasts: func(room *application.ChatRoom, frozenTime time.Time) []Broadcast {
				return []Broadcast{
					{
						room: room,
						msg: &domain.UserLeftSystemMessage{
							Timestamp: frozenTime,
							Name:      "Jane Doe",
						},
					},
					{
						room: room,
						msg: &domain.UserLeftSystemMessage{
							Timestamp: frozenTime,
							Name:      "John Doe",
						},
					},
				}
			},
			expectedClientsCount: 0,
		},
		{
			name: "success with two users entering and one leaving",
			clientsIn: []struct{ id, name string }{
				{"1", "Jane Doe"},
				{"2", "John Doe"},
			},
			clientsOut: []string{"2"},
			expectedBroadcasts: func(room *application.ChatRoom, frozenTime time.Time) []Broadcast {
				return []Broadcast{
					{
						room: room,
						msg: &domain.UserLeftSystemMessage{
							Timestamp: frozenTime,
							Name:      "John Doe",
						},
					},
				}
			},
			expectedClientsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			clientRegistry := application.NewClientRegistry()
			room := application.NewChatRoom("1", "general", clientRegistry)
			frozenTime := time.Date(2025, 12, 7, 0, 0, 0, 0, time.UTC)

			spyNotifier := SpyNotifier{broadcasts: make([]Broadcast, 0)}
			chatService := application.NewChatService(room, &spyNotifier, func() time.Time { return frozenTime }, 3, 3)

			for _, client := range tt.clientsIn {
				room.LetClientIn(domain.NewClient(client.id, client.name))
			}

			chatService.Start(ctx)

			for _, client := range tt.clientsOut {
				chatService.LeaveRoom(client)
			}

			time.Sleep(50 * time.Millisecond)

			expectedBroadcasts := tt.expectedBroadcasts(room, frozenTime)
			assert.Equal(t, len(expectedBroadcasts), len(spyNotifier.broadcasts))

			for idx, broadcast := range spyNotifier.broadcasts {
				userLeft, ok := broadcast.msg.(*domain.UserLeftSystemMessage)
				if !ok {
					t.Errorf("expected a user left message, got %T", broadcast)
				}
				assert.Equal(t, expectedBroadcasts[idx].msg, userLeft)
				assert.Equal(t, expectedBroadcasts[idx].room, broadcast.room)
			}

			assert.Len(t, room.GetClients(), tt.expectedClientsCount)
		})
	}
}
