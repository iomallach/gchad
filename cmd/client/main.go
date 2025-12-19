package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/client/application"
	"github.com/iomallach/gchad/internal/client/domain"
	"github.com/iomallach/gchad/internal/client/infrastructure"
	"github.com/iomallach/gchad/internal/client/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger := infrastructure.NewZeroLogLogger(log.Logger)
	dialer := infrastructure.NewWebsocketDialer(websocket.DefaultDialer, nil)
	url := infrastructure.NewUrl(
		"ws",
		"localhost",
		"chat",
		8080,
		infrastructure.NewQueryParam("name", ""),
	)
	chatClient := infrastructure.NewChatClient(
		dialer,
		make(chan application.Message, 256),
		make(chan domain.ChatMessage, 256),
		make(chan error, 256),
		logger,
		url,
	)
	login := ui.InitialLoginModel("Who are you?", ui.DefaultLoginScreenKeymap, chatClient)
	chat := ui.InitialChatModel(ui.DefaultChatScreenKeymap, chatClient)
	// TODO: pass a logger instead of nils
	model := ui.InitialAppModel(login, chat, chatClient)
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Printf("there has been an error: %s", err.Error())
	}
}
