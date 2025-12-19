package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iomallach/gchad/internal/client/application"
	"github.com/iomallach/gchad/internal/client/domain"
)

// Catppuccin Mocha colors
var (
	timestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#b4befe"))              // Lavender
	nameStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)   // Blue
	textStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))              // Text
	systemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Italic(true) // Yellow
)

type ChatScreenKeymap struct {
	CtrlC key.Binding
	Enter key.Binding
	Esc   key.Binding
	CtrlD key.Binding
}

var DefaultChatScreenKeymap = ChatScreenKeymap{
	CtrlC: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "toggle viewport/input"),
	),
	CtrlD: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "disconnect"),
	),
}

type newMessageReceived struct {
	msg application.Message
}

type newErrorReceived struct {
	err error
}

func pollForChatMessageCmd(chatClient application.ChatClient) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-chatClient.InboundMessages():
			return newMessageReceived{msg}
		case err := <-chatClient.Errors():
			return newErrorReceived{err}
		}
	}
}

type Chat struct {
	input        textinput.Model
	chatViewPort viewport.Model
	ready        bool
	bindings     ChatScreenKeymap
	inputFocused bool
	chatClient   application.ChatClient
	// TODO: this has to be a struct or something that manages the chat properly?
	messages []string
}

func InitialChatModel(bindings ChatScreenKeymap, chatClient application.ChatClient) Chat {
	return Chat{bindings: bindings, inputFocused: true, chatClient: chatClient, messages: make([]string, 0)}
}

func (c Chat) Init() tea.Cmd {
	return nil
}

func (c Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		c, cmd = c.updateOnWindowSizeChange(msg)

	case tea.KeyMsg:
		if key.Matches(msg, c.bindings.CtrlD) {
			_ = c.chatClient.Disconnect() // swallow the error?
			return c, func() tea.Msg {
				return disconnected{}
			}
		}
		if c.inputFocused {
			switch {
			case key.Matches(msg, c.bindings.CtrlC):
				_ = c.chatClient.Disconnect() // swallow the error, we're quitting
				return c, tea.Quit

			case key.Matches(msg, c.bindings.Enter):
				if c.input.Value() != "" {
					go c.chatClient.SendMessage(c.input.Value())
					c.input.Reset()
				}
				return c, nil

			case key.Matches(msg, c.bindings.Esc):
				c.inputFocused = false
				c.input.Blur()

			default:
				c.input, cmd = c.input.Update(msg)
			}
		} else {
			switch {
			case key.Matches(msg, c.bindings.Esc):
				c.inputFocused = true
				c.input.Focus()
			case key.Matches(msg, c.bindings.CtrlC):
				return c, tea.Quit
			}
		}

	case newMessageReceived:
		c.updateMessages(msg.msg)
		c.chatViewPort.SetContent(strings.Join(c.messages, "\n"))

		return c, pollForChatMessageCmd(c.chatClient)

	case switchToChat:
		return c, pollForChatMessageCmd(c.chatClient)
	}

	return c, cmd
}

func (c Chat) updateOnWindowSizeChange(msg tea.WindowSizeMsg) (Chat, tea.Cmd) {
	if !c.ready {
		c.input = textinput.New()
		c.input.Width = msg.Width - 2
		c.input.Focus()
		c.chatViewPort = viewport.New(msg.Width-2, msg.Height-2)

		c.ready = true
	} else {
		c.chatViewPort.Width = msg.Width - 2
		c.chatViewPort.Height = msg.Height - 2
		c.input.Width = msg.Width - 2
	}

	return c, nil
}

func (c *Chat) updateMessages(msg application.Message) {
	switch msg := msg.(type) {
	case domain.ChatMessage:
		time := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
		name := nameStyle.Render(msg.From + ":")
		text := textStyle.Render(msg.Text)
		c.messages = append(c.messages, fmt.Sprintf("%s %s %s", time, name, text))

	case domain.UserJoinedMessage:
		time := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
		text := systemStyle.Render(msg.Name + " joined!")
		c.messages = append(c.messages, fmt.Sprintf("%s %s", time, text))

	case domain.UserLeftMessage:
		time := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
		text := systemStyle.Render(msg.Name + " left!")
		c.messages = append(c.messages, fmt.Sprintf("%s %s", time, text))
	}
}

func (c Chat) View() string {
	return fmt.Sprintf("%s\n\n%s", c.chatViewPort.View(), c.input.View())
}
