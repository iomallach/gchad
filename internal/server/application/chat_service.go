package application

import (
	"context"

	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/iomallach/gchad/pkg/logging"
)

type ChatServicer interface {
	EnterRoom(clientId string, clientName string)
	LeaveRoom(clientId string)
	SendMessage(clientId string, msg string)
}

type ChatService struct {
	room     *ChatRoom // TODO: later needs to be a repository of chat rooms
	events   chan domain.ApplicationEvent
	messages chan *domain.UserMessage
	notifier Notifier
	clock    ClockGen
	logger   logging.Logger
}

func NewChatService(room *ChatRoom, notifier Notifier, clock ClockGen, eventsChanSize int, messagesChanSize int, logger logging.Logger) *ChatService {
	return &ChatService{
		room:     room,
		events:   make(chan domain.ApplicationEvent, eventsChanSize),
		messages: make(chan *domain.UserMessage, messagesChanSize),
		notifier: notifier,
		clock:    clock,
		logger:   logger,
	}
}

func (cs *ChatService) Start(ctx context.Context) {
	go cs.handleEvents(ctx)
	go cs.handleMessages(ctx)
}

func (cs *ChatService) EnterRoom(clientId string, clientName string) {
	event := cs.room.LetClientIn(domain.NewClient(clientId, clientName))

	select {
	case cs.events <- event:
	default:
		cs.logger.Error("event channel full", make(map[string]any))
	}
}

func (cs *ChatService) LeaveRoom(clientId string) {
	event := cs.room.LetClientOut(clientId)

	select {
	case cs.events <- event:
	default:
		cs.logger.Error("event channel full", make(map[string]any))
	}
}

func (cs *ChatService) SendMessage(clientId string, msg string) {
	userMessage := domain.NewUserMessage(msg, cs.clock(), clientId)

	select {
	case cs.messages <- userMessage:
	default:
		cs.logger.Error("message channel full", make(map[string]any))
	}
}

func (cs *ChatService) handleMessages(ctx context.Context) {
	for {
		select {
		case msg := <-cs.messages:
			cs.notifier.BroadcastToRoom(cs.room, msg)
		case <-ctx.Done():
			cs.logger.Info("message handler context done, exiting", make(map[string]any))
			return
		}
	}
}

func (cs *ChatService) handleEvents(ctx context.Context) {
	for {
		select {
		case event := <-cs.events:
			switch e := event.(type) {
			case *domain.UserJoinedRoom:
				joinedMsg := domain.NewUserJoinedSystemMessage(e.Name, cs.clock())
				cs.notifier.BroadcastToRoom(cs.room, joinedMsg)

			case *domain.UserLeftRoom:
				leftMessage := domain.NewUserLeftSystemMessage(e.Name, cs.clock())
				cs.notifier.BroadcastToRoom(cs.room, leftMessage)
			}
		case <-ctx.Done():
			cs.logger.Info("event handler context done, exiting", make(map[string]any))
			return
		}
	}
}
