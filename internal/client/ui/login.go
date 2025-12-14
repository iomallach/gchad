package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Login struct {
	textAboveInput string
	input          textinput.Model
}

func InitialLoginModel(textAboveInput string) Login {
	input := textinput.New()
	input.CharLimit = 20
	input.Width = 40
	input.Placeholder = "Username"
	input.Validate = func(string) error { return nil }

	return Login{
		textAboveInput: textAboveInput,
		input:          input,
	}
}

func (l Login) Init() tea.Cmd {
	return nil
}

func (l Login) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		l.input.Focus()
		switch msg.String() {
		case "ctrl+c":
			return l, tea.Quit
		case "enter":
			value := l.input.Value()
			l.textAboveInput = fmt.Sprintf("going to connect as %s", value)
			l.input.Reset()
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
