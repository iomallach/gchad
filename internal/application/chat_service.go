package application

import "github.com/iomallach/gchad/internal/domain"

type Notifier interface {
	BroadcastToRoom(*ChatRoom, domain.Messager) error
}

type ChatService struct {
	room     *ChatRoom // TODO: later needs to be a repository of chat rooms
	events   chan domain.ApplicationEvent
	notifier Notifier
	clock    ClockGen
}

func NewChatService(room *ChatRoom, notifier Notifier, clock ClockGen, eventsChanSize int) *ChatService {
	return &ChatService{
		room:     room,
		events:   make(chan domain.ApplicationEvent, eventsChanSize),
		notifier: notifier,
		clock:    clock,
	}
}

func (cs *ChatService) Start() {
	go cs.handleEvents()
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

	return cs.notifier.BroadcastToRoom(cs.room, userMessage)
}

func (cs *ChatService) handleEvents() {
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
		default:
		}
	}
}
