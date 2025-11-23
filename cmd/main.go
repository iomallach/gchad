package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader websocket.Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	// TODO: probably shouldnt be the actual hub here, just the broadcast channel
	hub  *Hub
	conn *websocket.Conn
	// TODO: clients don't need the full Message with the type, the Messager interface should be well enough
	send chan *gchad.Message
	id   string
	name string
}

func (client *Client) readLoop() {
	defer func() {
		client.hub.unregister <- client
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("client_id", client.id).Msg("unexpected websocket close")
			}
			break
		}

		messager, err := gchad.UnmarshalMessage(message)
		if err != nil {
			log.Error().Err(err).Str("client_id", client.id).Msg("failed to unmarshal message")
			continue
		}

		log.Debug().Str("client_id", client.id).Str("type", string(messager.MessageType())).Msg("received message")

		message_data, err := json.Marshal(messager)
		if err != nil {
			log.Error().Err(err).Str("client_id", client.id).Msg("failed to marshal message data")
		}
		broadcast_message := gchad.Message{
			MessageType: messager.MessageType(),
			Data:        message_data,
		}

		client.hub.broadcast <- &broadcast_message
	}
}

func (client *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
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

			log.Debug().Str("client_id", client.id).Str("type", string(message.MessageType)).Msg("sending message")

			json_message, err := json.Marshal(message)
			if err != nil {
				log.Error().Err(err).Str("client_id", client.id).Msg("failed to marshal message")
			}

			err = client.conn.WriteMessage(websocket.TextMessage, json_message)
			if err != nil {
				log.Error().Err(err).Str("client_id", client.id).Msg("failed to write message")
				continue
			}

			log.Debug().Str("client_id", client.id).Msg("message sent")
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type Hub struct {
	broadcast  chan *gchad.Message
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
}

func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client.id] = client
			log.Info().Str("client_id", client.id).Str("name", client.name).Int("total_clients", len(hub.clients)).Msg("client registered")

			msg, err := json.Marshal(gchad.UserJoinedSystemMessage{
				Timestamp: time.Now(),
				Name:      client.name,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal user joined message")
				continue
			}
			for _, client := range hub.clients {
				client.send <- &gchad.Message{
					MessageType: gchad.SystemUserJoined,
					Data:        msg,
				}
			}
		case client := <-hub.unregister:
			if _, ok := hub.clients[client.id]; ok {
				delete(hub.clients, client.id)
				close(client.send)
				log.Info().Str("client_id", client.id).Str("name", client.name).Int("total_clients", len(hub.clients)).Msg("client unregistered")

				msg, err := json.Marshal(gchad.UserLeftSystemMessage{
					Timestamp: time.Now(),
					Name:      client.name,
				})
				if err != nil {
					log.Error().Err(err).Msg("failed to marshal user left message")
					continue
				}
				for _, client := range hub.clients {
					client.send <- &gchad.Message{
						MessageType: gchad.SystemUserLeft,
						Data:        msg,
					}
				}
			}
		case message := <-hub.broadcast:
			log.Debug().Str("type", string(message.MessageType)).Int("recipients", len(hub.clients)).Msg("broadcasting message")
			for _, client := range hub.clients {
				client.send <- message
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("remote_addr", r.RemoteAddr).Msg("new websocket connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	body := make([]byte, 256)
	read_bytes, err := r.Body.Read(body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read request body")
	}

	msg, err := gchad.UnmarshalMessage(body[:read_bytes])
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal connect message")
	}

	connectMeMessage, ok := msg.(*gchad.ConnectMeMessage)
	if !ok {
		log.Error().Msg("expected ConnectMeMessage, got different type")
	} else {

		client := &Client{hub: hub, conn: conn, send: make(chan *gchad.Message, 256), id: uuid.NewString(), name: connectMeMessage.Name}
		hub.register <- client

		log.Debug().Str("client_id", client.id).Str("name", client.name).Msg("starting client loops")
		go client.readLoop()
		go client.writeLoop()
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	hub := Hub{
		broadcast:  make(chan *gchad.Message),
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()
	log.Info().Msg("hub started")

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		serveWs(&hub, w, r)
	})

	log.Info().Str("addr", ":8080").Msg("server starting")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
