package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/client/infrastructure"
	"github.com/iomallach/gchad/internal/client/ui"
)

func main() {
	login := ui.InitialLoginModel("Who are you?", ui.DefaultLoginScreenKeymap)
	chat := ui.InitialChatModel(ui.DefaultChatScreenKeymap)
	// TODO: pass a logger instead of nils
	dialer := infrastructure.NewWebsocketDialer(websocket.DefaultDialer, nil)
	chatClient := infrastructure.NewChatClient(
		dialer,
		make(chan any, 256),
		make(chan any, 256),
		make(chan error, 256),
		nil,
	)
	model := ui.InitialAppModel(login, chat, chatClient)
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Printf("there has been an error: %s", err.Error())
	}
}
