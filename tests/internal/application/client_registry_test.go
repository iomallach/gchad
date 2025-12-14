package application_test

import (
	"testing"

	"github.com/iomallach/gchad/internal/server/application"
	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestClientRegistry_AddClient_GetClient(t *testing.T) {
	registry := application.NewClientRegistry()
	client := domain.NewClient("1", "Jane Doe")

	registry.AddClient(client)

	assert.Equal(t, client, registry.GetClient(client.Id()))
}

func TestClientRegistry_RemoveClient(t *testing.T) {
	registry := application.NewClientRegistry()
	client := domain.NewClient("1", "Jane Doe")

	registry.AddClient(client)

	assert.Equal(t, client, registry.GetClient(client.Id()))

	registry.RemoveClient(client.Id())

	assert.Equal(t, []*domain.Client{}, registry.GetAllClients())
}

func TestClientRegistry_GetAllClients(t *testing.T) {
	registry := application.NewClientRegistry()
	client1 := domain.NewClient("1", "Jane Doe")
	client2 := domain.NewClient("2", "John Doe")

	registry.AddClient(client1)
	registry.AddClient(client2)

	clients := registry.GetAllClients()
	assert.Len(t, clients, 2)
	assert.Contains(t, clients, client1)
	assert.Contains(t, clients, client2)
}
