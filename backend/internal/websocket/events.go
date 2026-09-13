package websocket

import (
	"time"

	"media-sequencer-backend/internal/models"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	EventTypeSyncStart       EventType = "SYNC_START"
	EventTypeSyncEnd         EventType = "SYNC_END"
	EventTypePlaylistUpdated EventType = "PLAYLIST_UPDATED"
	EventTypeHeartbeat       EventType = "HEARTBEAT"
)

// BaseEvent is the standard wrapper for all WS messages
type BaseEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

// SyncStartPayload is broadcast when a sync event is triggered
type SyncStartPayload struct {
	SyncID    string        `json:"syncId"`
	MediaID   int           `json:"mediaId"`
	Duration  int           `json:"duration"` // Duration in seconds
	StartAt   time.Time     `json:"startAt"`
	EndAt     time.Time     `json:"endAt"`
	Media     *models.Media `json:"media"`
}

// SyncEndPayload is broadcast when a sync event ends
type SyncEndPayload struct {
	SyncID string `json:"syncId"`
}

// PlaylistUpdatedPayload is broadcast when a window playlist changes
type PlaylistUpdatedPayload struct {
	WindowID int `json:"windowId"`
}
