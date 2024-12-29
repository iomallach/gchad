package main

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	send chan []byte
	id   string
}

func (client *Client) readLoop() {
	defer func() {
		client.hub.unregister <- client
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		log.Println("Received message:", string(message))
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		client.hub.broadcast <- message
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
			log.Println("[Write loop] Client: ", client.id, ", sending message:", string(message))
			if !ok {
				// The hub closed the channel
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = writer.Write(message)
			if err != nil {
				return
			}

			for i := 0; i < len(client.send); i++ {
				_, err = writer.Write([]byte{'\n'})
				if err != nil {
					return
				}
				_, err = writer.Write(<-client.send)
				if err != nil {
					return
				}
			}
			if err := writer.Close(); err != nil {
				return
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
	broadcast  chan []byte
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
			// TODO: notify all a client connected
		case client := <-hub.unregister:
			if _, ok := hub.clients[client.id]; ok {
				delete(hub.clients, client.id)
				close(client.send)
				// TODO: notify all a client disconnected
			}
		case message := <-hub.broadcast:
			log.Println("Broadcasting message: ", string(message))
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

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), id: uuid.NewString()}
	hub.register <- client

	log.Println("Spawning read/write loops")
	go client.readLoop()
	go client.writeLoop()
}

func main() {
	hub := Hub{
		broadcast:  make(chan []byte),
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
