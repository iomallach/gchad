package gchad

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	TimeFunc func() time.Time
	Hub      struct {
		Broadcast  chan *Message
		clients    map[string]*Client
		register   chan *Client
		Unregister chan *Client
		timeFunc   TimeFunc
	}
)

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client.Id] = client
			log.Info().Str("client_id", client.Id).Str("name", client.Name).Int("total_clients", len(hub.clients)).Msg("client registered")

			msg, err := json.Marshal(UserJoinedSystemMessage{
				Timestamp: time.Now(),
				Name:      client.Name,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal user joined message")
				continue
			}
			for _, client := range hub.clients {
				client.send <- &Message{
					MessageType: SystemUserJoined,
					Data:        msg,
				}
			}
		case client := <-hub.Unregister:
			if _, ok := hub.clients[client.Id]; ok {
				delete(hub.clients, client.Id)
				close(client.send)
				log.Info().Str("client_id", client.Id).Str("name", client.Name).Int("total_clients", len(hub.clients)).Msg("client unregistered")

				msg, err := json.Marshal(UserLeftSystemMessage{
					Timestamp: time.Now(),
					Name:      client.Name,
				})
				if err != nil {
					log.Error().Err(err).Msg("failed to marshal user left message")
					continue
				}
				for _, client := range hub.clients {
					client.send <- &Message{
						MessageType: SystemUserLeft,
						Data:        msg,
					}
				}
			}
		case message := <-hub.Broadcast:
			log.Debug().Str("type", string(message.MessageType)).Int("recipients", len(hub.clients)).Msg("broadcasting message")
			for _, client := range hub.clients {
				client.send <- message
			}
		}
	}
}

func (hub *Hub) ScheduleRegisterClient(client *Client) {
	hub.register <- client
}

func NewHub(broadcast chan *Message, register chan *Client, unregister chan *Client, timeFunc TimeFunc) Hub {
	return Hub{
		Broadcast:  broadcast,
		clients:    make(map[string]*Client),
		register:   register,
		Unregister: unregister,
		timeFunc:   timeFunc,
	}
}
