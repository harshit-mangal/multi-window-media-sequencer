package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"media-sequencer-backend/internal/models"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 512),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	log.Println("[WEBSOCKET_HUB] WebSocket Hub is running")
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			total := len(h.clients)
			h.mu.Unlock()
			log.Printf("[WEBSOCKET_CONNECTED] Client connected. Total active clients: %d", total)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			total := len(h.clients)
			h.mu.Unlock()
			log.Printf("[WEBSOCKET_DISCONNECTED] Client disconnected. Total active clients: %d", total)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow client buffer full, drop and cleanup
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastJSON serializes data and sends to all clients
func (h *Hub) BroadcastJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[WEBSOCKET_ERROR] Failed to marshal broadcast message: %v", err)
		return
	}
	h.broadcast <- data
}

// BroadcastSyncStart sends a SYNC_START event
func (h *Hub) BroadcastSyncStart(syncID string, mediaID int, duration int, startAt, endAt time.Time, media *models.Media) {
	event := BaseEvent{
		Type:      EventTypeSyncStart,
		Timestamp: time.Now(),
		Payload: SyncStartPayload{
			SyncID:   syncID,
			MediaID:  mediaID,
			Duration: duration,
			StartAt:  startAt,
			EndAt:    endAt,
			Media:    media,
		},
	}
	log.Printf("[SYNC_STARTED] Broadcasting SYNC_START for MediaID=%d, Duration=%ds, StartAt=%s, EndAt=%s",
		mediaID, duration, startAt.Format(time.RFC3339), endAt.Format(time.RFC3339))
	h.BroadcastJSON(event)
}

// BroadcastSyncEnd sends a SYNC_END event
func (h *Hub) BroadcastSyncEnd(syncID string) {
	event := BaseEvent{
		Type:      EventTypeSyncEnd,
		Timestamp: time.Now(),
		Payload: SyncEndPayload{
			SyncID: syncID,
		},
	}
	log.Printf("[SYNC_FINISHED] Broadcasting SYNC_END for SyncID=%s", syncID)
	h.BroadcastJSON(event)
}

// BroadcastPlaylistUpdated sends a PLAYLIST_UPDATED event
func (h *Hub) BroadcastPlaylistUpdated(windowID int) {
	event := BaseEvent{
		Type:      EventTypePlaylistUpdated,
		Timestamp: time.Now(),
		Payload: PlaylistUpdatedPayload{
			WindowID: windowID,
		},
	}
	log.Printf("[PLAYLIST_UPDATED] Broadcasting PLAYLIST_UPDATED for WindowID=%d", windowID)
	h.BroadcastJSON(event)
}

// ActiveClientCount returns the number of active connected clients
func (h *Hub) ActiveClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
