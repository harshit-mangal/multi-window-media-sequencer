package models

import (
	"time"
)

// MediaType represents supported media types
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeBlank MediaType = "blank"
)

// Window represents a physical or logical display window
type Window struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Playlist  []PlaylistItem `json:"playlist,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Media represents a media asset (image, video, or explicit blank)
type Media struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      MediaType `json:"type"`
	URL       string    `json:"url"`
	Duration  int       `json:"duration"` // Duration in seconds
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PlaylistItem represents an ordered media entry in a window's playlist
type PlaylistItem struct {
	ID        int       `json:"id"`
	WindowID  int       `json:"windowId"`
	MediaID   int       `json:"mediaId"`
	Position  int       `json:"position"`
	Media     *Media    `json:"media,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SyncEvent represents a temporary playback synchronization override
type SyncEvent struct {
	ID        string    `json:"id"`
	MediaID   int       `json:"mediaId"`
	Duration  int       `json:"duration"` // Duration in seconds
	StartAt   time.Time `json:"startAt"`
	EndAt     time.Time `json:"endAt"`
	Status    string    `json:"status"` // "active", "completed", "cancelled"
	Media     *Media    `json:"media,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// PlaybackState represents computed deterministic playback position for a window
type PlaybackState struct {
	WindowID            int           `json:"windowId"`
	WindowName          string        `json:"windowName"`
	CycleNumber         int64         `json:"cycleNumber"`
	CycleElapsedSec     int           `json:"cycleElapsedSec"`
	PlaylistTotalSec    int           `json:"playlistTotalSec"`
	CurrentItemIndex    int           `json:"currentItemIndex"`
	CurrentPlaylistItem *PlaylistItem `json:"currentPlaylistItem,omitempty"`
	ItemElapsedSec      int           `json:"itemElapsedSec"`
	ItemRemainingSec    int           `json:"itemRemainingSec"`
	NextPlaylistItem    *PlaylistItem `json:"nextPlaylistItem,omitempty"`
	ServerTime          time.Time     `json:"serverTime"`
	IsSyncOverride      bool          `json:"isSyncOverride"`
	ActiveSyncEvent     *SyncEvent    `json:"activeSyncEvent,omitempty"`
}

// DTO Requests and Responses

type CreateMediaRequest struct {
	Name     string    `json:"name" binding:"required,min=1,max=150"`
	Type     MediaType `json:"type" binding:"required,oneof=image video blank"`
	URL      string    `json:"url" binding:"required"`
	Duration int       `json:"duration" binding:"required,gt=0"`
}

type AddToPlaylistRequest struct {
	MediaID  int  `json:"mediaId" binding:"required,gt=0"`
	Position *int `json:"position,omitempty"` // If null, append to end
}

type UpdatePlaylistItemRequest struct {
	Position *int `json:"position,omitempty"`
	MediaID  *int `json:"mediaId,omitempty"`
}

type StartSyncRequest struct {
	MediaID  int `json:"mediaId" binding:"required,gt=0"`
	Duration int `json:"duration" binding:"required,gt=0,lte=18000"` // In seconds (up to 5 hours max)
}

type SyncCurrentResponse struct {
	Active       bool       `json:"active"`
	SyncEvent    *SyncEvent `json:"syncEvent,omitempty"`
	RemainingSec int        `json:"remainingSec,omitempty"`
	ServerTime   time.Time  `json:"serverTime"`
}

// Structured error response
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
