package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/server/application"
	"github.com/iomallach/gchad/internal/server/infrastructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := infrastructure.NewZeroLogLogger(log.Logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	room := application.NewChatRoom("1", "General", application.NewClientRegistry())
	// TODO: maybe the notifier shouldn't be exposed here at all, and shall handle
	// registration calls via chat service telling it to do so?
	notifier := infrastructure.NewClientNotifier(logger, make(map[string]*infrastructure.Client))
	chatService := application.NewChatService(room, notifier, func() time.Time { return time.Now() }, 256, 256, logger)

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	clientConfig := infrastructure.ClientConfiguration{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      (60 * 9 * time.Second) / 10,
		RecieveChanWait: 10 * time.Second,
		SendChannelSize: 256,
		RecvChannelSize: 256,
	}
	handler := infrastructure.NewHandler(
		upgrader,
		chatService,
		notifier,
		clientConfig,
		func() string { return uuid.NewString() },
		logger,
		ctx,
	)

	chatService.Start(ctx)

	http.HandleFunc("/chat", handler.ServeHTTP)

	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		logger.Info("server starting, serving the chat at /chat", map[string]any{"port": "8080"})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", map[string]any{"error": err.Error()})
		}
	}()

	<-sigChan
	logger.Info("Received signal, shutting down", map[string]any{})
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", map[string]any{"error": err.Error()})
	}

	logger.Info("server stopped", map[string]any{})
}
