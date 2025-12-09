package application

import (
	"context"
	"fmt"

	"github.com/iomallach/gchad/internal/domain"
)

type ChatService struct {
	room     *ChatRoom // TODO: later needs to be a repository of chat rooms
	events   chan domain.ApplicationEvent
	messages chan *domain.UserMessage
	notifier Notifier
	clock    ClockGen
	logger   Logger
}

func NewChatService(room *ChatRoom, notifier Notifier, clock ClockGen, eventsChanSize int, messagesChanSize int, logger Logger) *ChatService {
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
	userMessage := domain.NewUserMessage(msg, cs.clock())

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
			if err := cs.notifier.BroadcastToRoom(cs.room, msg); err != nil {
				cs.logger.Error(fmt.Sprintf("error broadcasting to room: %s", err.Error()), make(map[string]any))
			}
		case <-ctx.Done():
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
				if err := cs.notifier.BroadcastToRoom(cs.room, joinedMsg); err != nil {
					cs.logger.Error(fmt.Sprintf("error broadcasting to room: %s", err.Error()), make(map[string]any))
				}
			case *domain.UserLeftRoom:
				leftMessage := domain.NewUserLeftSystemMessage(e.Name, cs.clock())
				if err := cs.notifier.BroadcastToRoom(cs.room, leftMessage); err != nil {
					cs.logger.Error(fmt.Sprintf("error broadcasting to room: %s", err.Error()), make(map[string]any))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
