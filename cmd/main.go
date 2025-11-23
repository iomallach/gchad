package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad"
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
				log.Printf("error: %v", err)
			}
			break
		}

		messager, err := gchad.UnmarshalMessage(message)
		if err != nil {
			log.Printf("Error unmarshalling message: %s", err.Error())
			continue
		}

		log.Printf("Received message of type: %s", messager.MessageType())

		message_data, err := json.Marshal(messager)
		if err != nil {
			log.Printf("Error marshalling message data: %s", err.Error())
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
				// The hub closed the channel
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Println("[Write loop] Client: ", client.id, ", sending message of type:", message.MessageType)

			json_message, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("Error marshalling a message at client %s", client.id)
			}

			err = client.conn.WriteMessage(websocket.TextMessage, json_message)
			if err != nil {
				fmt.Printf("Error writing a message at client %s", client.id)
				continue
			}

			log.Println("Message sent")
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
			log.Println("Registered client: ", client.id)

			msg, err := json.Marshal(gchad.UserJoinedSystemMessage{
				Timestamp: time.Now(),
				Name:      client.name,
			})
			if err != nil {
				log.Printf("Error marshalling user joined message")
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

				msg, err := json.Marshal(gchad.UserLeftSystemMessage{
					Timestamp: time.Now(),
					Name:      client.name,
				})
				if err != nil {
					log.Printf("Error marshalling user left message")
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
			log.Println("Broadcasting message of type: ", message.MessageType)
			for _, client := range hub.clients {
				client.send <- message
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Println("Received new connection from: ", r.Host)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	body := make([]byte, 256)
	read_bytes, err := r.Body.Read(body)
	if err != nil {
		log.Printf("Failed to read body: %s", err.Error())
	}

	msg, err := gchad.UnmarshalMessage(body[:read_bytes])
	if err != nil {
		log.Printf("Failed to unmarshal message: %s", err.Error())
	}

	connectMeMessage, ok := msg.(*gchad.ConnectMeMessage)
	if !ok {
		log.Printf("Failed to read connect me message")
	} else {

		client := &Client{hub: hub, conn: conn, send: make(chan *gchad.Message, 256), id: uuid.NewString(), name: connectMeMessage.Name}
		hub.register <- client

		log.Println("Spawning read/write loops")
		go client.readLoop()
		go client.writeLoop()
	}
}

func main() {
	hub := Hub{
		broadcast:  make(chan *gchad.Message),
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()
	log.Println("Starting hub goroutine")

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		serveWs(&hub, w, r)
	})

	log.Println("Starting server at localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
