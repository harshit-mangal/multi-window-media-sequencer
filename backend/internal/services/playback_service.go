package services

import (
	"context"
	"time"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
)

type PlaybackService interface {
	ComputePlaybackState(ctx context.Context, window *models.Window, playlist []models.PlaylistItem, now time.Time) *models.PlaybackState
	ComputeAllPlaybackStates(ctx context.Context, windows []models.Window, now time.Time) ([]models.PlaybackState, error)
}

type playbackService struct {
	cfg      *config.Config
	syncRepo repository.SyncRepository
}

func NewPlaybackService(cfg *config.Config, syncRepo repository.SyncRepository) PlaybackService {
	return &playbackService{
		cfg:      cfg,
		syncRepo: syncRepo,
	}
}

// ComputePlaybackState determines the exact media item playing in a 5-hour cycle.
// This calculation is purely deterministic and derived from:
// 1. Fixed cycle duration (18,000s = 5 hours)
// 2. Cumulative playlist durations
// 3. Absolute server timestamp
// 4. Active temporary synchronization overrides
func (s *playbackService) ComputePlaybackState(ctx context.Context, window *models.Window, playlist []models.PlaylistItem, now time.Time) *models.PlaybackState {
	serverUnix := now.Unix()
	cycleDuration := int64(s.cfg.CycleDurationSec) // 18000 seconds = 5 hours

	// Calculate current 5-hour cycle number and elapsed seconds in this cycle
	// Using UNIX epoch as anchor ensures consistent calculation across all servers & clients
	cycleNumber := serverUnix / cycleDuration
	cycleElapsedSec := int(serverUnix % cycleDuration)

	state := &models.PlaybackState{
		WindowID:        window.ID,
		WindowName:      window.Name,
		CycleNumber:     cycleNumber,
		CycleElapsedSec: cycleElapsedSec,
		ServerTime:      now,
	}

	// Check if there is an active temporary sync event
	if s.syncRepo != nil {
		activeSync, err := s.syncRepo.GetCurrentActive(ctx, now)
		if err == nil && activeSync != nil {
			// Check if current time is within [startAt, endAt] or in countdown buffer
			if !now.Before(activeSync.StartAt) && now.Before(activeSync.EndAt) {
				state.IsSyncOverride = true
				state.ActiveSyncEvent = activeSync
			}
		}
	}

	// Calculate total playlist duration
	totalPlaylistSec := 0
	for _, item := range playlist {
		if item.Media != nil && item.Media.Duration > 0 {
			totalPlaylistSec += item.Media.Duration
		}
	}
	state.PlaylistTotalSec = totalPlaylistSec

	if len(playlist) == 0 || totalPlaylistSec == 0 {
		return state
	}

	// Calculate loop position within playlist for the current cycle
	// When playlist finishes, it repeats from the beginning up until the 5-hour cycle boundary.
	// At cycle boundary, cycle N ends and cycle N+1 begins.
	playLoopElapsedSec := cycleElapsedSec % totalPlaylistSec

	// Find the active item by accumulating durations
	accumulatedSec := 0
	for i, item := range playlist {
		itemDuration := 0
		if item.Media != nil {
			itemDuration = item.Media.Duration
		}
		if itemDuration <= 0 {
			continue
		}

		if playLoopElapsedSec >= accumulatedSec && playLoopElapsedSec < accumulatedSec+itemDuration {
			// Found current item
			itemCopy := item
			state.CurrentItemIndex = i
			state.CurrentPlaylistItem = &itemCopy
			state.ItemElapsedSec = playLoopElapsedSec - accumulatedSec
			state.ItemRemainingSec = itemDuration - state.ItemElapsedSec

			// Determine next item (wraps to beginning of playlist)
			nextIndex := (i + 1) % len(playlist)
			nextItemCopy := playlist[nextIndex]
			state.NextPlaylistItem = &nextItemCopy
			break
		}
		accumulatedSec += itemDuration
	}

	return state
}

func (s *playbackService) ComputeAllPlaybackStates(ctx context.Context, windows []models.Window, now time.Time) ([]models.PlaybackState, error) {
	states := make([]models.PlaybackState, 0, len(windows))
	for _, w := range windows {
		state := s.ComputePlaybackState(ctx, &w, w.Playlist, now)
		states = append(states, *state)
	}
	return states, nil
}
