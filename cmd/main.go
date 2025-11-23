package main

import (
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

func serveWs(hub *gchad.Hub, w http.ResponseWriter, r *http.Request) {
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

		client := gchad.NewClient(hub, conn, make(chan *gchad.Message, 256), uuid.NewString(), connectMeMessage.Name, pingPeriod, writeWait)
		hub.ScheduleRegisterClient(&client)

		log.Debug().Str("client_id", client.Id).Str("name", client.Name).Msg("starting client loops")
		go client.ReadLoop()
		go client.WriteLoop()
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	hub := gchad.NewHub(make(chan *gchad.Message), make(chan *gchad.Client), make(chan *gchad.Client), func() time.Time {
		return time.Now()
	})
	go hub.Run()
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
