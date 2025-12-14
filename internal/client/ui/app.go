package ui

import tea "github.com/charmbracelet/bubbletea"

type loginSucceeded struct {
	name string
}

type disconnected struct{}

type activeScreen int

const (
	loginScreen = iota
	chatScreen
)

type screenSize struct {
	height int
	width  int
}

type App struct {
	loginScreen  Login
	chatScreen   Chat
	screenSize   screenSize
	activeScreen activeScreen
	name         string
}

func InitialAppModel() App {
	return App{
		loginScreen:  InitialLoginModel("Who are you?", DefaultLoginScreenKeymap),
		chatScreen:   InitialChatModel(DefaultChatScreenKeymap),
		activeScreen: loginScreen,
		screenSize:   screenSize{},
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
		a.activeScreen = chatScreen
		a.name = msg.name
		return a, func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.screenSize.width, Height: a.screenSize.height}
		}

	case disconnected:
		a.activeScreen = loginScreen
		a.loginScreen = InitialLoginModel("Who are you?", DefaultLoginScreenKeymap)
		a.name = ""
		return a, nil
	}

	switch a.activeScreen {
	case loginScreen:
		updatedLoginScreen, cmd := a.loginScreen.Update(msg)
		a.loginScreen = updatedLoginScreen.(Login)
		return a, cmd
	case chatScreen:
		updatedChatScreen, cmd := a.chatScreen.Update(msg)
		a.chatScreen = updatedChatScreen.(Chat)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.activeScreen {
	case loginScreen:
		return a.loginScreen.View()
	case chatScreen:
		return a.chatScreen.View()
	default:
		panic("unknown active screen")
	}
}
