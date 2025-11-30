package gchad

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	SetWriteDeadline(time.Time) error
}

type HubNotifier interface {
	UnregisterClient(client *Client)
	SendToBroadcast(message *Message)
}

type ChannelHubNotifier struct {
	unregister chan *Client
	broadcast  chan *Message
}

func NewChannelHubNotifier(unregister chan *Client, broadcast chan *Message) *ChannelHubNotifier {
	return &ChannelHubNotifier{
		unregister,
		broadcast,
	}
}

func (notifier *ChannelHubNotifier) UnregisterClient(client *Client) {
	notifier.unregister <- client
}

func (notifier *ChannelHubNotifier) SendToBroadcast(message *Message) {
	notifier.broadcast <- message
}

type Client struct {
	// TODO: probably shouldnt be the actual hub here, just the broadcast channel
	notifier HubNotifier
	conn     Connection
	// TODO: clients don't need the full Message with the type, the Messager interface should be well enough
	send       chan *Message
	Id         string
	Name       string
	pingPeriod time.Duration
	writeWait  time.Duration
}

func NewClient(notifier HubNotifier, conn Connection, send chan *Message, id string, name string, pingPeriod time.Duration, writeWait time.Duration) Client {
	return Client{
		notifier:   notifier,
		conn:       conn,
		send:       send,
		Id:         id,
		Name:       name,
		pingPeriod: pingPeriod,
		writeWait:  writeWait,
	}
}

func (client *Client) ReadLoop() {
	defer func() {
		client.notifier.UnregisterClient(client)
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Str("client_id", client.Id).Msg("could not read message")
			break
		}

		messager, err := UnmarshalMessage(message)
		if err != nil {
			log.Error().Err(err).Str("client_id", client.Id).Msg("failed to unmarshal message")
			continue
		}

		log.Debug().Str("client_id", client.Id).Str("type", string(messager.MessageType())).Msg("received message")

		message_data, err := json.Marshal(messager)
		if err != nil {
			log.Error().Err(err).Str("client_id", client.Id).Msg("failed to marshal message data")
		}
		broadcast_message := Message{
			MessageType: messager.MessageType(),
			Data:        message_data,
		}

		client.notifier.SendToBroadcast(&broadcast_message)
	}
}

func (client *Client) WriteLoop() {
	ticker := time.NewTicker(client.pingPeriod)
	defer func() {
		ticker.Stop()
		close(client.send)
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Debug().Str("client_id", client.Id).Str("type", string(message.MessageType)).Msg("sending message")

			json_message, err := json.Marshal(message)
			if err != nil {
				log.Error().Err(err).Str("client_id", client.Id).Msg("failed to marshal message")
			}

			err = client.conn.WriteMessage(websocket.TextMessage, json_message)
			if err != nil {
				log.Error().Err(err).Str("client_id", client.Id).Msg("failed to write message")
				continue
			}

			log.Debug().Str("client_id", client.Id).Msg("message sent")
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(client.writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
