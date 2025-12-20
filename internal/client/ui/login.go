package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iomallach/gchad/internal/client/application"
)

var rootStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	Padding(2)

var textAboveStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF5F87")).
	Bold(true)

type failedToConnectToChat struct {
	err error
}

func connectToChatCmd(chatClient application.ChatClient, name string) tea.Cmd {
	return func() tea.Msg {
		chatClient.SetName(name)
		if err := chatClient.Connect(); err != nil {
			return failedToConnectToChat{err}
		}

		return switchToChat{name}
	}
}

type LoginScreenKeymap struct {
	CtrlC key.Binding
	Enter key.Binding
}

func (km *LoginScreenKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.CtrlC, km.Enter},
	}
}

var DefaultLoginScreenKeymap = LoginScreenKeymap{
	CtrlC: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
}

type Login struct {
	chatClient     application.ChatClient
	textAboveInput string
	input          textinput.Model
	bindings       LoginScreenKeymap
	width          int
	height         int
}

func InitialLoginModel(textAboveInput string, bindings LoginScreenKeymap, chatClient application.ChatClient) Login {
	input := textinput.New()
	input.CharLimit = 20
	input.Width = 40
	input.Placeholder = "Username"
	input.Validate = func(s string) error {
		if strings.Contains(s, " ") || strings.Contains(s, "\n") || strings.Contains(s, "\t") {
			return fmt.Errorf("username cannot contain spaces")
		}
		return nil
	}
	input.Focus()

	return Login{
		textAboveInput: textAboveInput,
		input:          input,
		bindings:       bindings,
		chatClient:     chatClient,
	}
}

func (l Login) Init() tea.Cmd {
	return nil
}

func (l Login) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height

		return l, nil

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, l.bindings.CtrlC):
			return l, tea.Quit

		case key.Matches(msg, l.bindings.Enter):
			name := l.input.Value()
			err := l.input.Validate(name)

			if err != nil {
				l.textAboveInput = fmt.Sprintf("invalid username: %s", err.Error())
			} else {
				l.textAboveInput = fmt.Sprintf("going to connect as %s", name)
				l.input.Reset()

				return l, connectToChatCmd(l.chatClient, name)
			}

			return l, nil
		default:
			var cmd tea.Cmd
			l.input, cmd = l.input.Update(msg)

			return l, cmd

		}

	case failedToConnectToChat:
		// TODO:display the error somewhere? Popup? Press any key to continue?
		l.textAboveInput = fmt.Sprintf("failed to connect: %s", msg.err.Error())

		return l, nil
	}

	return l, nil
}

func (l Login) View() string {
	styledText := textAboveStyle.Render(l.textAboveInput)
	content := fmt.Sprintf("%s\n\n%s", styledText, l.input.View())
	box := rootStyle.Render(content)

	return lipgloss.Place(l.width, l.height/3, lipgloss.Center, lipgloss.Center, box)
}
