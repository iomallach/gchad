package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iomallach/gchad/internal/client/domain"
)

// Catppuccin Mocha colors for statusline
var (
	statusGreen   = lipgloss.Color("#a6e3a1") // Green
	statusBase    = lipgloss.Color("#1e1e2e") // Base background
	statusSurface = lipgloss.Color("#313244") // Surface0
	statusText    = lipgloss.Color("#cdd6f4") // Text

	statusLeftStyle = lipgloss.NewStyle().
			Background(statusGreen).
			Foreground(statusBase).
			Bold(true).
			Padding(0, 1)

	statusMiddleStyle = lipgloss.NewStyle().
				Foreground(statusText).
				Background(statusGreen)

	statusRightStyle = lipgloss.NewStyle().
				Background(statusGreen).
				Foreground(statusBase).
				Bold(true).
				Padding(0, 1)

	statusRootStyle = lipgloss.NewStyle().
			Background(statusSurface)

	leftSectionRightSeparatorStyle = lipgloss.NewStyle().
					Foreground(statusGreen).
					Background(statusSurface)

	middleSectionSeparatorStyle = lipgloss.NewStyle().
					Foreground(statusGreen).
					Background(statusSurface)

	rightSeparatorStyle = lipgloss.NewStyle().
				Foreground(statusGreen).
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
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

func (s StatusLine) View() string {
	// Left section: Connected as info
	leftSection := statusLeftStyle.Render(fmt.Sprintf("As %s", s.connectedAs))
	leftSectionRightSeparator := leftSectionRightSeparatorStyle.Render("") // U+E0B0: right-pointing triangle

	// Middle section: Stats
	middleSection := statusMiddleStyle.Render(fmt.Sprintf(" In: %d Out: %d ", s.messagesReceived, s.messagesSent))
	middleSectionLeftSeparator := middleSectionSeparatorStyle.Render("")
	middleSectionRightSeparator := middleSectionSeparatorStyle.Render("")

	// Right section: gchad
	rightSeparator := rightSeparatorStyle.Render("") // U+E0B2: left-pointing triangle
	rightSection := statusRightStyle.Render("gchad")

	// Calculate content widths (using lipgloss.Width to account for ANSI codes)
	leftWidth := lipgloss.Width(leftSection) + lipgloss.Width(leftSectionRightSeparator)
	middleWidth := lipgloss.Width(middleSection) + lipgloss.Width(middleSectionLeftSeparator) + lipgloss.Width(middleSectionRightSeparator)
	rightWidth := lipgloss.Width(rightSeparator) + lipgloss.Width(rightSection)

	// Calculate spacing needed
	if s.width > 0 {
		totalContentWidth := leftWidth + middleWidth + rightWidth
		remainingSpace := s.width - totalContentWidth

		if remainingSpace > 0 {
			// Split remaining space: half before middle, half after middle
			leftPadding := remainingSpace / 2
			rightPadding := remainingSpace - leftPadding

			// Create padding with root background
			leftGap := statusRootStyle.Render(lipgloss.NewStyle().Width(leftPadding).Render(""))
			rightGap := statusRootStyle.Render(lipgloss.NewStyle().Width(rightPadding).Render(""))

			return statusRootStyle.Render(
				leftSection +
					leftSectionRightSeparator +
					leftGap +
					middleSectionLeftSeparator +
					middleSection +
					middleSectionRightSeparator +
					rightGap +
					rightSeparator +
					rightSection,
			)
		}
	}

	// Fallback if width not set or content too wide
	return statusRootStyle.Render(
		leftSection +
			leftSectionRightSeparator + " " +
			middleSectionLeftSeparator +
			middleSection + " " +
			rightSeparator +
			rightSection,
	)
}
