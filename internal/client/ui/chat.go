package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iomallach/gchad/internal/client/application"
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
	msg any
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
	// TODO: this has to be a struct or something that manages the chat properly
	messages []any
}

func InitialChatModel(bindings ChatScreenKeymap, chatClient application.ChatClient) Chat {
	return Chat{bindings: bindings, inputFocused: true, chatClient: chatClient, messages: make([]any, 0)}
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
			return c, func() tea.Msg {
				return disconnected{}
			}
		}
		if c.inputFocused {
			switch {
			case key.Matches(msg, c.bindings.CtrlC):
				return c, tea.Quit
			case key.Matches(msg, c.bindings.Enter):
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
		c.messages = append(c.messages, msg.msg)
		// Then rendering takes care of it?

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
		c.chatViewPort.SetContent("There is some initial context/nAnd some more")

		c.ready = true
	} else {
		c.chatViewPort.Width = msg.Width - 2
		c.chatViewPort.Height = msg.Height - 2
		c.input.Width = msg.Width - 2
	}

	return c, nil
}

func (c Chat) View() string {
	return fmt.Sprintf("%s\n\n%s", c.chatViewPort.View(), c.input.View())
}
