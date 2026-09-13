package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "media-sequencer-backend/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow cross-origin WebSocket connections
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWS upgrades HTTP connection to WebSocket and registers client
// GET /ws
func (h *WebSocketHandler) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WEBSOCKET_ERROR] Failed to upgrade connection: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn)
	h.hub.Register <- client

	// Start client read and write pumps in goroutines
	go client.WritePump()
	go client.ReadPump()
}
