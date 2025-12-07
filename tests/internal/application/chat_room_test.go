package application_test

import (
	"testing"

	"github.com/iomallach/gchad/internal/application"
	"github.com/iomallach/gchad/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestChatRoom_LetClientIn(t *testing.T) {
	chatRoom := application.NewChatRoom("1", "general", application.NewClientRegistry())
	client := domain.NewClient("1", "Jane Doe")
	expectedEvent := domain.NewUserJoinedRoomEvent("Jane Doe")

	event := chatRoom.LetClientIn(client)
	clients := chatRoom.GetClients()

	assert.NotNil(t, event)
	assert.Equal(t, expectedEvent, event)
	assert.Equal(t, []*domain.Client{client}, clients)
}

func TestChatRoom_LetClientOut(t *testing.T) {
	chatRoom := application.NewChatRoom("1", "general", application.NewClientRegistry())
	client := domain.NewClient("1", "Jane Doe")

	event := chatRoom.LetClientIn(client)
	assert.NotNil(t, event)

	chatRoom.LetClientOut(client.Id())

	assert.Equal(t, 0, len(chatRoom.GetClients()))
}
