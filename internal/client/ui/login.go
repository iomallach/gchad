package ui

import (
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
	return l, nil
}

func (l Login) View() string {
	return ""
}
