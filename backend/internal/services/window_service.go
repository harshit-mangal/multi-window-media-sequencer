package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
)

type WindowService interface {
	GetAllWindows(ctx context.Context) ([]models.Window, error)
	GetWindowByID(ctx context.Context, id int) (*models.Window, *models.PlaybackState, error)
	CreateWindow(ctx context.Context, name string) (*models.Window, error)
}

type windowService struct {
	windowRepo      repository.WindowRepository
	playlistRepo    repository.PlaylistRepository
	playbackService PlaybackService
}

func NewWindowService(
	windowRepo repository.WindowRepository,
	playlistRepo repository.PlaylistRepository,
	playbackService PlaybackService,
) WindowService {
	return &windowService{
		windowRepo:      windowRepo,
		playlistRepo:    playlistRepo,
		playbackService: playbackService,
	}
}

func (s *windowService) GetAllWindows(ctx context.Context) ([]models.Window, error) {
	windows, err := s.windowRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch windows: %w", err)
	}

	for i := range windows {
		playlist, err := s.playlistRepo.GetByWindowID(ctx, windows[i].ID)
		if err == nil {
			windows[i].Playlist = playlist
		}
	}

	return windows, nil
}

func (s *windowService) GetWindowByID(ctx context.Context, id int) (*models.Window, *models.PlaybackState, error) {
	window, err := s.windowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get window: %w", err)
	}
	if window == nil {
		return nil, nil, errors.New("window not found")
	}

	playlist, err := s.playlistRepo.GetByWindowID(ctx, window.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get window playlist: %w", err)
	}
	window.Playlist = playlist

	playbackState := s.playbackService.ComputePlaybackState(ctx, window, playlist, time.Now().UTC())
	return window, playbackState, nil
}

func (s *windowService) CreateWindow(ctx context.Context, name string) (*models.Window, error) {
	if name == "" {
		return nil, errors.New("window name cannot be empty")
	}
	return s.windowRepo.Create(ctx, name)
}
