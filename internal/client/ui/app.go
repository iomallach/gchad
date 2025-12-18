package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iomallach/gchad/internal/client/application"
)

type switchToChat struct{}

type disconnected struct{}

type activeSession int

const (
	loginSession = iota
	chatSession
)

type screenSize struct {
	height int
	width  int
}

type App struct {
	loginScreen   Login
	chatScreen    Chat
	screenSize    screenSize
	activeSession activeSession
}

func InitialAppModel(loginScreen Login, chatScreen Chat, chatClient application.ChatClient) App {
	return App{
		loginScreen:   loginScreen,
		chatScreen:    chatScreen,
		activeSession: loginSession,
		screenSize:    screenSize{},
	}
}
func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.screenSize.width = msg.Width
		a.screenSize.height = msg.Height
		updatedLoginScreen, loginUpdCmd := a.loginScreen.Update(msg)
		updatedChatScreen, chatUpdCmd := a.chatScreen.Update(msg)
		a.loginScreen = updatedLoginScreen.(Login)
		a.chatScreen = updatedChatScreen.(Chat)

		return a, tea.Batch(loginUpdCmd, chatUpdCmd)

	case switchToChat:
		// this activates the switch below and delivers the message over to the chat screen
		a.activeSession = chatSession

	case disconnected:
		a.activeSession = loginSession

	}

	switch a.activeSession {
	case loginSession:
		updatedLoginScreen, cmd := a.loginScreen.Update(msg)
		a.loginScreen = updatedLoginScreen.(Login)
		return a, cmd
	case chatSession:
		updatedChatScreen, cmd := a.chatScreen.Update(msg)
		a.chatScreen = updatedChatScreen.(Chat)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.activeSession {
	case loginSession:
		return a.loginScreen.View()
	case chatSession:
		return a.chatScreen.View()
	default:
		panic("unknown active screen")
	}
}
