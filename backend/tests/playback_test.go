package tests

import (
	"context"
	"testing"
	"time"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
)

func TestPlaybackCalculation_EmptyPlaylist(t *testing.T) {
	cfg := &config.Config{
		CycleDurationSec: 18000,
	}
	playbackSvc := services.NewPlaybackService(cfg, nil)

	window := &models.Window{ID: 1, Name: "Window 1"}
	playlist := []models.PlaylistItem{}

	now := time.Unix(100, 0).UTC()
	state := playbackSvc.ComputePlaybackState(context.Background(), window, playlist, now)

	if state.PlaylistTotalSec != 0 {
		t.Fatalf("expected PlaylistTotalSec=0, got %d", state.PlaylistTotalSec)
	}
	if state.CurrentPlaylistItem != nil {
		t.Fatalf("expected CurrentPlaylistItem=nil for empty playlist, got %+v", state.CurrentPlaylistItem)
	}
}

func TestPlaybackCalculation_SingleItemLoop(t *testing.T) {
	cfg := &config.Config{
		CycleDurationSec: 18000,
	}
	playbackSvc := services.NewPlaybackService(cfg, nil)

	window := &models.Window{ID: 1, Name: "Window 1"}
	media1 := models.Media{ID: 1, Name: "M1", Type: models.MediaTypeImage, Duration: 10}
	playlist := []models.PlaylistItem{
		{ID: 101, WindowID: 1, MediaID: 1, Position: 0, Media: &media1},
	}

	// At t = 0
	state0 := playbackSvc.ComputePlaybackState(context.Background(), window, playlist, time.Unix(0, 0).UTC())
	if state0.CurrentItemIndex != 0 || state0.ItemElapsedSec != 0 || state0.ItemRemainingSec != 10 {
		t.Fatalf("at t=0, unexpected state: index=%d elapsed=%d remaining=%d", state0.CurrentItemIndex, state0.ItemElapsedSec, state0.ItemRemainingSec)
	}

	// At t = 5
	state5 := playbackSvc.ComputePlaybackState(context.Background(), window, playlist, time.Unix(5, 0).UTC())
	if state5.ItemElapsedSec != 5 || state5.ItemRemainingSec != 5 {
		t.Fatalf("at t=5, unexpected state: elapsed=%d remaining=%d", state5.ItemElapsedSec, state5.ItemRemainingSec)
	}

	// At t = 10 (loop 1 starts)
	state10 := playbackSvc.ComputePlaybackState(context.Background(), window, playlist, time.Unix(10, 0).UTC())
	if state10.ItemElapsedSec != 0 || state10.ItemRemainingSec != 10 {
		t.Fatalf("at t=10, unexpected loop restart: elapsed=%d remaining=%d", state10.ItemElapsedSec, state10.ItemRemainingSec)
	}
}

func TestPlaybackCalculation_MultiItemSequenceAnd5HourCycle(t *testing.T) {
	cfg := &config.Config{
		CycleDurationSec: 18000, // 5 hours
	}
	playbackSvc := services.NewPlaybackService(cfg, nil)

	window := &models.Window{ID: 1, Name: "Window 1"}
	m1 := models.Media{ID: 1, Name: "M1", Type: models.MediaTypeImage, Duration: 10}
	m2 := models.Media{ID: 2, Name: "M2", Type: models.MediaTypeImage, Duration: 20}
	m3 := models.Media{ID: 3, Name: "M3", Type: models.MediaTypeVideo, Duration: 30}

	// Total duration = 10 + 20 + 30 = 60s
	playlist := []models.PlaylistItem{
		{ID: 1, WindowID: 1, MediaID: 1, Position: 0, Media: &m1},
		{ID: 2, WindowID: 1, MediaID: 2, Position: 1, Media: &m2},
		{ID: 3, WindowID: 1, MediaID: 3, Position: 2, Media: &m3},
	}

	testCases := []struct {
		timestampSec   int64
		expectedIndex  int
		expectedMedia  int
		expectedElapsed int
		expectedRemain int
		expectedNext   int
		expectedCycle  int64
	}{
		// [0, 10) -> M1
		{timestampSec: 0, expectedIndex: 0, expectedMedia: 1, expectedElapsed: 0, expectedRemain: 10, expectedNext: 2, expectedCycle: 0},
		{timestampSec: 5, expectedIndex: 0, expectedMedia: 1, expectedElapsed: 5, expectedRemain: 5, expectedNext: 2, expectedCycle: 0},
		// [10, 30) -> M2
		{timestampSec: 10, expectedIndex: 1, expectedMedia: 2, expectedElapsed: 0, expectedRemain: 20, expectedNext: 3, expectedCycle: 0},
		{timestampSec: 25, expectedIndex: 1, expectedMedia: 2, expectedElapsed: 15, expectedRemain: 5, expectedNext: 3, expectedCycle: 0},
		// [30, 60) -> M3
		{timestampSec: 30, expectedIndex: 2, expectedMedia: 3, expectedElapsed: 0, expectedRemain: 30, expectedNext: 1, expectedCycle: 0},
		{timestampSec: 59, expectedIndex: 2, expectedMedia: 3, expectedElapsed: 29, expectedRemain: 1, expectedNext: 1, expectedCycle: 0},
		// [60, 70) -> M1 (Loop 1)
		{timestampSec: 60, expectedIndex: 0, expectedMedia: 1, expectedElapsed: 0, expectedRemain: 10, expectedNext: 2, expectedCycle: 0},
		{timestampSec: 75, expectedIndex: 1, expectedMedia: 2, expectedElapsed: 5, expectedRemain: 15, expectedNext: 3, expectedCycle: 0},
		// End of 5-hour cycle 0 (t = 17999) -> 17999 % 60 = 59 (M3 at 29s elapsed)
		{timestampSec: 17999, expectedIndex: 2, expectedMedia: 3, expectedElapsed: 29, expectedRemain: 1, expectedNext: 1, expectedCycle: 0},
		// Start of 5-hour cycle 1 (t = 18000) -> 18000 % 18000 = 0 -> M1 at 0s elapsed
		{timestampSec: 18000, expectedIndex: 0, expectedMedia: 1, expectedElapsed: 0, expectedRemain: 10, expectedNext: 2, expectedCycle: 1},
		// Cycle 1 at t = 18015 -> 18015 % 18000 = 15s -> 15 % 60 = 15 -> M2 at 5s elapsed
		{timestampSec: 18015, expectedIndex: 1, expectedMedia: 2, expectedElapsed: 5, expectedRemain: 15, expectedNext: 3, expectedCycle: 1},
	}

	for _, tc := range testCases {
		now := time.Unix(tc.timestampSec, 0).UTC()
		state := playbackSvc.ComputePlaybackState(context.Background(), window, playlist, now)

		if state.CycleNumber != tc.expectedCycle {
			t.Errorf("t=%d: expected CycleNumber=%d, got %d", tc.timestampSec, tc.expectedCycle, state.CycleNumber)
		}
		if state.CurrentItemIndex != tc.expectedIndex {
			t.Errorf("t=%d: expected CurrentItemIndex=%d, got %d", tc.timestampSec, tc.expectedIndex, state.CurrentItemIndex)
		}
		if state.CurrentPlaylistItem == nil || state.CurrentPlaylistItem.MediaID != tc.expectedMedia {
			t.Errorf("t=%d: expected MediaID=%d, got %+v", tc.timestampSec, tc.expectedMedia, state.CurrentPlaylistItem)
		}
		if state.ItemElapsedSec != tc.expectedElapsed {
			t.Errorf("t=%d: expected ItemElapsedSec=%d, got %d", tc.timestampSec, tc.expectedElapsed, state.ItemElapsedSec)
		}
		if state.ItemRemainingSec != tc.expectedRemain {
			t.Errorf("t=%d: expected ItemRemainingSec=%d, got %d", tc.timestampSec, tc.expectedRemain, state.ItemRemainingSec)
		}
		if state.NextPlaylistItem == nil || state.NextPlaylistItem.MediaID != tc.expectedNext {
			t.Errorf("t=%d: expected NextMediaID=%d, got %+v", tc.timestampSec, tc.expectedNext, state.NextPlaylistItem)
		}
	}
}
