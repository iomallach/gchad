package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iomallach/gchad/internal/client/ui"
)

func main() {
	model := ui.InitialAppModel()
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Printf("there has been an error: %s", err.Error())
	}
}
