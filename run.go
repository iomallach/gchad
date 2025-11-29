package gchad

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type UUIDMaker func() string

func MapConnectMeMessage(messageBody []byte) (*ConnectMeMessage, error) {
	msg, err := UnmarshalMessage(messageBody)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal connect message")
		return nil, err
	}

	connectMeMessage, ok := msg.(*ConnectMeMessage)
	if !ok {
		log.Error().Msg("expected ConnectMeMessage, got different type")
		return nil, fmt.Errorf("could not convert interface to type")
	}

	return connectMeMessage, nil
}

func CreateClient(message *ConnectMeMessage, hub *Hub, conn Connection, sendChannelSize int, uuidMaker UUIDMaker, pingPeriod time.Duration, writePeriod time.Duration) Client {
	return NewClient(hub, conn, make(chan *Message, sendChannelSize), uuidMaker(), message.Name, pingPeriod, writePeriod)
}

func LauchClient(hub *Hub, client *Client) {
	hub.ScheduleRegisterClient(client)

	log.Debug().Str("client_id", client.Id).Str("name", client.Name).Msg("starting client loops")
	go client.ReadLoop()
	go client.WriteLoop()
}
