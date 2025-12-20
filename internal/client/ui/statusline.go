package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iomallach/gchad/internal/client/domain"
)

var (
	statusLeftStyle = lipgloss.NewStyle().
			Background(CatppuccinMocha.Green).
			Foreground(CatppuccinMocha.Base).
			Bold(true).
			Padding(0, 1)

	statusMiddleStyle = lipgloss.NewStyle().
				Foreground(CatppuccinMocha.Text).
				Background(CatppuccinMocha.Crust)

	statusRightStyle = lipgloss.NewStyle().
				Background(CatppuccinMocha.Green).
				Foreground(CatppuccinMocha.Base).
				Bold(true).
				Padding(0, 1)

	statusRootStyle = lipgloss.NewStyle().
			Background(CatppuccinMocha.Surface0)

	leftSectionRightSeparatorStyle = lipgloss.NewStyle().
					Foreground(CatppuccinMocha.Green).
					Background(CatppuccinMocha.Surface0)

	middleSectionSeparatorStyle = lipgloss.NewStyle().
					Foreground(CatppuccinMocha.Crust).
					Background(CatppuccinMocha.Surface0)

	rightSectionLeftSeparatorStyle = lipgloss.NewStyle().
					Foreground(CatppuccinMocha.Green).
					Background(CatppuccinMocha.Surface0)
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
	leftSection := statusLeftStyle.Render(fmt.Sprintf(" %s", s.connectedAs))
	leftSectionRightSeparator := leftSectionRightSeparatorStyle.Render("") // U+E0B0: right-pointing triangle

	middleSection := statusMiddleStyle.Render(fmt.Sprintf(" In: %d Out: %d Online: %d ", s.messagesReceived, s.messagesSent, s.clientsInTheRoom))
	middleSectionLeftSeparator := middleSectionSeparatorStyle.Render("")
	middleSectionRightSeparator := middleSectionSeparatorStyle.Render("")

	rightSectionLeftSeparator := rightSectionLeftSeparatorStyle.Render("") // U+E0B2: left-pointing triangle
	rightSection := statusRightStyle.Render(" gchad")

	leftWidth := lipgloss.Width(leftSection) + lipgloss.Width(leftSectionRightSeparator)
	middleWidth := lipgloss.Width(middleSection) + lipgloss.Width(middleSectionLeftSeparator) + lipgloss.Width(middleSectionRightSeparator)
	rightWidth := lipgloss.Width(rightSectionLeftSeparator) + lipgloss.Width(rightSection)

	// spacing required
	if s.width > 0 {
		totalContentWidth := leftWidth + middleWidth + rightWidth
		remainingSpace := s.width - totalContentWidth

		if remainingSpace > 0 {
			// split remaining space: half before middle, half after middle
			leftPadding := remainingSpace / 2
			rightPadding := remainingSpace - leftPadding

			// create padding with root background
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
					rightSectionLeftSeparator +
					rightSection,
			)
		}
	}

	// fallback if width not set or content too wide
	return statusRootStyle.Render(
		leftSection +
			leftSectionRightSeparator + " " +
			middleSectionLeftSeparator +
			middleSection + " " +
			rightSectionLeftSeparator +
			rightSection,
	)
}
