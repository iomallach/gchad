package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/client/domain"
	"github.com/iomallach/gchad/internal/client/infrastructure"
	"github.com/iomallach/gchad/internal/client/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Redirect logs to a file to avoid breaking the TUI
	logFile, err := os.OpenFile("client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger := infrastructure.NewZeroLogLogger(log.Logger)
	dialer := infrastructure.NewWebsocketDialer(websocket.DefaultDialer, logger)
	url := infrastructure.NewUrl(
		"ws",
		"localhost",
		"chat",
		8080,
		infrastructure.NewQueryParam("name", ""),
	)
	communications := infrastructure.NewCommunications(
		make(chan domain.Message, 256),
		make(chan domain.ChatMessage, 256),
		make(chan error, 256),
	)
	chatClient := infrastructure.NewChatClient(
		dialer,
		communications,
		logger,
		url,
	)
	login := ui.InitialLoginModel("Who are you?", ui.DefaultLoginScreenKeymap, chatClient)
	chat := ui.InitialChatModel(ui.DefaultChatScreenKeymap, chatClient, ui.NewMessageRingBuffer(100))
	model := ui.InitialAppModel(login, chat, chatClient)
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Printf("there has been an error: %s", err.Error())
	}
}
