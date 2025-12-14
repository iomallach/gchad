package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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
	textAboveInput string
	input          textinput.Model
	bindings       LoginScreenKeymap
}

func InitialLoginModel(textAboveInput string, bindings LoginScreenKeymap) Login {
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

	return Login{
		textAboveInput: textAboveInput,
		input:          input,
		bindings:       bindings,
	}
}

func (l Login) Init() tea.Cmd {
	return nil
}

func (l Login) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		l.input.Focus()
		switch {
		case key.Matches(msg, l.bindings.CtrlC):
			return l, tea.Quit
		case key.Matches(msg, l.bindings.Enter):
			value := l.input.Value()
			err := l.input.Validate(value)
			if err != nil {
				l.textAboveInput = fmt.Sprintf("invalid username: %s", err.Error())
			} else {
				l.textAboveInput = fmt.Sprintf("going to connect as %s", value)
				l.input.Reset()
			}
			return l, nil
		}
	}

	var cmd tea.Cmd
	l.input, cmd = l.input.Update(msg)

	return l, cmd
}

func (l Login) View() string {
	return fmt.Sprintf("%s\n\n%s", l.textAboveInput, l.input.View())
}
