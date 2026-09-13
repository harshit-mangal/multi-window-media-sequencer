package services

import (
	"context"
	"errors"
	"fmt"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
	"media-sequencer-backend/internal/websocket"
)

type PlaylistService interface {
	GetPlaylist(ctx context.Context, windowID int) ([]models.PlaylistItem, error)
	AddItem(ctx context.Context, windowID int, req models.AddToPlaylistRequest) (*models.PlaylistItem, error)
	UpdateItem(ctx context.Context, windowID int, itemID int, req models.UpdatePlaylistItemRequest) (*models.PlaylistItem, error)
	RemoveItem(ctx context.Context, windowID int, itemID int) error
}

type playlistService struct {
	playlistRepo repository.PlaylistRepository
	windowRepo   repository.WindowRepository
	mediaRepo    repository.MediaRepository
	wsHub        *websocket.Hub
}

func NewPlaylistService(
	playlistRepo repository.PlaylistRepository,
	windowRepo repository.WindowRepository,
	mediaRepo repository.MediaRepository,
	wsHub *websocket.Hub,
) PlaylistService {
	return &playlistService{
		playlistRepo: playlistRepo,
		windowRepo:   windowRepo,
		mediaRepo:    mediaRepo,
		wsHub:        wsHub,
	}
}

func (s *playlistService) GetPlaylist(ctx context.Context, windowID int) ([]models.PlaylistItem, error) {
	window, err := s.windowRepo.GetByID(ctx, windowID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify window: %w", err)
	}
	if window == nil {
		return nil, errors.New("window not found")
	}

	return s.playlistRepo.GetByWindowID(ctx, windowID)
}

func (s *playlistService) AddItem(ctx context.Context, windowID int, req models.AddToPlaylistRequest) (*models.PlaylistItem, error) {
	// 1. Verify window exists
	window, err := s.windowRepo.GetByID(ctx, windowID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify window: %w", err)
	}
	if window == nil {
		return nil, errors.New("window not found")
	}

	// 2. Verify media exists
	media, err := s.mediaRepo.GetByID(ctx, req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify media: %w", err)
	}
	if media == nil {
		return nil, errors.New("media not found")
	}

	// 3. Add to repository
	item, err := s.playlistRepo.AddItem(ctx, windowID, req.MediaID, req.Position)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to playlist: %w", err)
	}

	// 4. Notify connected clients via WebSocket
	if s.wsHub != nil {
		s.wsHub.BroadcastPlaylistUpdated(windowID)
	}

	return item, nil
}

func (s *playlistService) UpdateItem(ctx context.Context, windowID int, itemID int, req models.UpdatePlaylistItemRequest) (*models.PlaylistItem, error) {
	// 1. Verify window exists
	window, err := s.windowRepo.GetByID(ctx, windowID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify window: %w", err)
	}
	if window == nil {
		return nil, errors.New("window not found")
	}

	// 2. If changing media, verify media exists
	if req.MediaID != nil {
		media, err := s.mediaRepo.GetByID(ctx, *req.MediaID)
		if err != nil || media == nil {
			return nil, errors.New("media not found")
		}
	}

	// 3. Update in repository
	item, err := s.playlistRepo.UpdateItem(ctx, itemID, req.Position, req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to update playlist item: %w", err)
	}

	// 4. Notify connected clients
	if s.wsHub != nil {
		s.wsHub.BroadcastPlaylistUpdated(windowID)
	}

	return item, nil
}

func (s *playlistService) RemoveItem(ctx context.Context, windowID int, itemID int) error {
	err := s.playlistRepo.RemoveItem(ctx, windowID, itemID)
	if err != nil {
		return fmt.Errorf("failed to remove playlist item: %w", err)
	}

	// Notify connected clients
	if s.wsHub != nil {
		s.wsHub.BroadcastPlaylistUpdated(windowID)
	}

	return nil
}
