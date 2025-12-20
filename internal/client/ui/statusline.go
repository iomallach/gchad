package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iomallach/gchad/internal/client/domain"
)

// Catppuccin Mocha colors for statusline
var (
	statusGreen      = lipgloss.Color("#a6e3a1") // Green
	statusBase       = lipgloss.Color("#1e1e2e") // Base background
	statusSurface    = lipgloss.Color("#313244") // Surface0
	statusText       = lipgloss.Color("#cdd6f4") // Text

	statusLeftStyle = lipgloss.NewStyle().
		Background(statusGreen).
		Foreground(statusBase).
		Bold(true).
		Padding(0, 1)

	statusMiddleStyle = lipgloss.NewStyle().
		Foreground(statusText)

	statusRightStyle = lipgloss.NewStyle().
		Background(statusGreen).
		Foreground(statusBase).
		Bold(true).
		Padding(0, 1)

	statusRootStyle = lipgloss.NewStyle().
		Background(statusSurface)
)

type StatusLine struct {
	room             string
	connectedAs      string
	messagesReceived int
	messagesSent     int
	clientsInTheRoom int
	width            int
}

func InitialStatusLineModel() StatusLine {
	return StatusLine{}
}

func (s StatusLine) Init() tea.Cmd {
	return nil
}

func (s StatusLine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *domain.ChatStats:
		s.clientsInTheRoom = msg.ClientsInTheRoom
		s.messagesReceived = msg.MessagesReceived
		s.messagesSent = msg.MessagesSent
	}
	return s, nil
}

func (s StatusLine) View() string {
	return fmt.Sprintf("As %s | In: %d Out: %d", s.connectedAs, s.messagesReceived, s.messagesSent)
}
