package application

import (
	"context"

	"github.com/iomallach/gchad/internal/domain"
)

type Notifier interface {
	BroadcastToRoom(*ChatRoom, domain.Messager) error
}

type ChatService struct {
	room     *ChatRoom // TODO: later needs to be a repository of chat rooms
	events   chan domain.ApplicationEvent
	messages chan *domain.UserMessage
	notifier Notifier
	clock    ClockGen
}

func NewChatService(room *ChatRoom, notifier Notifier, clock ClockGen, eventsChanSize int, messagesChanSize int) *ChatService {
	return &ChatService{
		room:     room,
		events:   make(chan domain.ApplicationEvent, eventsChanSize),
		messages: make(chan *domain.UserMessage, messagesChanSize),
		notifier: notifier,
		clock:    clock,
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
		// TODO: log event channel full
	}
}

func (cs *ChatService) LeaveRoom(clientId string) {
	event := cs.room.LetClientOut(clientId)

	select {
	case cs.events <- event:
	default:
		// TODO: log event channel full
	}
}

func (cs *ChatService) SendMessage(clientId string, msg string) error {
	userMessage := domain.NewUserMessage(msg, cs.clock())

	select {
	case cs.messages <- userMessage:
	default:
		// TODO: log message channel full
	}
	return cs.notifier.BroadcastToRoom(cs.room, userMessage)
}

func (cs *ChatService) handleMessages(ctx context.Context) {
	for {
		select {
		case msg := <-cs.messages:
			if err := cs.notifier.BroadcastToRoom(cs.room, msg); err != nil {
				// TODO: log error
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
					// TODO: log error
				}
			case *domain.UserLeftRoom:
				leftMessage := domain.NewUserLeftSystemMessage(e.Name, cs.clock())
				if err := cs.notifier.BroadcastToRoom(cs.room, leftMessage); err != nil {
					// TODO: log error
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
