package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iomallach/gchad/internal/client/application"
)

type loginSucceeded struct {
	name string
}

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
	name          string
	chatClient    application.ChatClient
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

	case loginSucceeded:
		a.activeSession = chatSession
		a.name = msg.name
		return a, func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.screenSize.width, Height: a.screenSize.height}
		}

	case disconnected:
		a.activeSession = loginSession
		a.loginScreen = InitialLoginModel("Who are you?", DefaultLoginScreenKeymap)
		a.name = ""
		return a, nil
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
