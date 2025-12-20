package infrastructure

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iomallach/gchad/internal/server/application"
	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/iomallach/gchad/pkg/logging"
	"github.com/iomallach/gchad/pkg/network"
)

type Handler struct {
	upgrader     websocket.Upgrader
	chatService  *application.ChatService
	notifier     *ClientNotifier
	clientConfig ClientConfiguration
	idGen        application.IdGen
	logger       logging.Logger
	appCtx       context.Context
}

func NewHandler(
	upgrader websocket.Upgrader,
	chatService *application.ChatService,
	notifier *ClientNotifier,
	clientConfig ClientConfiguration,
	idGen application.IdGen,
	logger logging.Logger,
	appCtx context.Context,
) *Handler {
	return &Handler{
		upgrader:     upgrader,
		chatService:  chatService,
		notifier:     notifier,
		clientConfig: clientConfig,
		idGen:        idGen,
		logger:       logger,
		appCtx:       appCtx,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientName := r.URL.Query().Get("name")
	if clientName == "" {
		h.logger.Error("client name not provided, skipping", map[string]any{})
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	h.logger.Info("upgraded connection to websocket", map[string]any{})
	if err != nil {
		h.logger.Error(fmt.Sprintf("upgrade failed: %s", err.Error()), map[string]any{})
		return
	}

	clientId := h.idGen()
	wsConn := network.NewWebsocketsConnection(conn, h.logger)
	recv := make(chan domain.Messager, h.clientConfig.RecvChannelSize)
	send := make(chan domain.Messager, h.clientConfig.SendChannelSize)
	client := NewClient(clientId, clientName, wsConn, recv, send, h.clientConfig, h.logger)
	h.logger.Info(fmt.Sprintf("client %s connected", clientName), map[string]any{})

	h.notifier.RegisterClient(client)
	h.chatService.EnterRoom(clientId, clientName)

	ctx, cancel := context.WithCancel(h.appCtx)

	go client.WriteMessages(ctx)
	go h.forwardMessages(ctx, clientName, recv)
	h.logger.Info(fmt.Sprintf("client %s started", clientName), map[string]any{})

	// TODO: blocks this goroutine, need to unblock it later
	client.ReadMessages(ctx)
	cancel()
	h.chatService.LeaveRoom(clientId)
	h.notifier.UnregisterClient(clientId)
}

func (h *Handler) forwardMessages(ctx context.Context, clientId string, recv chan domain.Messager) {
	for {
		select {
		case msg, ok := <-recv:
			if !ok {
				h.logger.Debug("client has closed, exiting forwardMessages", map[string]any{"client_id": clientId})
				return
			}
			if userMsg, ok := msg.(*domain.UserMessage); ok {
				h.chatService.SendMessage(clientId, userMsg.Text)
			}
		case <-ctx.Done():
			return
		}
	}
}
