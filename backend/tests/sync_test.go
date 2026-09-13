package tests

import (
	"context"
	"testing"
	"time"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
)

// MockSyncRepo provides an in-memory implementation of SyncRepository for isolated testing
type mockSyncRepo struct {
	events map[string]models.SyncEvent
}

func newMockSyncRepo() *mockSyncRepo {
	return &mockSyncRepo{events: make(map[string]models.SyncEvent)}
}

func (m *mockSyncRepo) Create(ctx context.Context, event models.SyncEvent) (*models.SyncEvent, error) {
	for k, v := range m.events {
		if v.Status == "active" {
			v.Status = "superseded"
			m.events[k] = v
		}
	}
	m.events[event.ID] = event
	return &event, nil
}

func (m *mockSyncRepo) GetCurrentActive(ctx context.Context, now time.Time) (*models.SyncEvent, error) {
	for _, e := range m.events {
		if e.Status == "active" && e.EndAt.After(now) {
			copy := e
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockSyncRepo) GetByID(ctx context.Context, id string) (*models.SyncEvent, error) {
	if e, ok := m.events[id]; ok {
		return &e, nil
	}
	return nil, nil
}

func (m *mockSyncRepo) ExpirePastEvents(ctx context.Context, now time.Time) error {
	for k, e := range m.events {
		if e.Status == "active" && (e.EndAt.Before(now) || e.EndAt.Equal(now)) {
			e.Status = "completed"
			m.events[k] = e
		}
	}
	return nil
}

type mockMediaRepo struct {
	media map[int]models.Media
}

func newMockMediaRepo() *mockMediaRepo {
	return &mockMediaRepo{
		media: map[int]models.Media{
			2: {ID: 2, Name: "Aurora Lights", Type: models.MediaTypeImage, Duration: 15},
		},
	}
}

func (m *mockMediaRepo) GetAll(ctx context.Context) ([]models.Media, error) {
	res := make([]models.Media, 0, len(m.media))
	for _, v := range m.media {
		res = append(res, v)
	}
	return res, nil
}

func (m *mockMediaRepo) GetByID(ctx context.Context, id int) (*models.Media, error) {
	if v, ok := m.media[id]; ok {
		return &v, nil
	}
	return nil, nil
}

func (m *mockMediaRepo) Create(ctx context.Context, req models.CreateMediaRequest) (*models.Media, error) {
	newMedia := models.Media{
		ID:       len(m.media) + 1,
		Name:     req.Name,
		Type:     req.Type,
		URL:      req.URL,
		Duration: req.Duration,
	}
	m.media[newMedia.ID] = newMedia
	return &newMedia, nil
}

func (m *mockMediaRepo) Delete(ctx context.Context, id int) error {
	delete(m.media, id)
	return nil
}

func TestSyncService_StartSyncAndActiveState(t *testing.T) {
	cfg := &config.Config{
		SyncBufferSec:    2,
		CycleDurationSec: 18000,
	}
	syncRepo := newMockSyncRepo()
	mediaRepo := newMockMediaRepo()

	syncSvc := services.NewSyncService(cfg, syncRepo, mediaRepo, nil)

	// Start sync for Media 2 with 20 seconds duration
	event, err := syncSvc.StartSync(context.Background(), models.StartSyncRequest{
		MediaID:  2,
		Duration: 20,
	})
	if err != nil {
		t.Fatalf("expected StartSync to succeed, got error: %v", err)
	}

	if event.MediaID != 2 {
		t.Errorf("expected MediaID=2, got %d", event.MediaID)
	}
	if event.Duration != 20 {
		t.Errorf("expected Duration=20, got %d", event.Duration)
	}
	if event.StartAt.After(event.EndAt) {
		t.Errorf("expected StartAt to be before EndAt")
	}

	// Verify current sync response
	current, err := syncSvc.GetCurrentSync(context.Background())
	if err != nil {
		t.Fatalf("expected GetCurrentSync to succeed, got error: %v", err)
	}

	if !current.Active {
		t.Errorf("expected active sync to be true")
	}
	if current.SyncEvent == nil || current.SyncEvent.MediaID != 2 {
		t.Errorf("expected syncEvent mediaID=2")
	}
}

func TestSyncOverride_NonDestructivePlayback(t *testing.T) {
	cfg := &config.Config{
		SyncBufferSec:    2,
		CycleDurationSec: 18000,
	}
	syncRepo := newMockSyncRepo()
	mediaRepo := newMockMediaRepo()
	syncSvc := services.NewSyncService(cfg, syncRepo, mediaRepo, nil)
	playbackSvc := services.NewPlaybackService(cfg, syncRepo)

	window := &models.Window{ID: 1, Name: "Window 1"}
	m1 := models.Media{ID: 1, Name: "M1", Type: models.MediaTypeImage, Duration: 10}
	m3 := models.Media{ID: 3, Name: "M3", Type: models.MediaTypeVideo, Duration: 20}
	originalPlaylist := []models.PlaylistItem{
		{ID: 1, WindowID: 1, MediaID: 1, Position: 0, Media: &m1},
		{ID: 3, WindowID: 1, MediaID: 3, Position: 1, Media: &m3},
	}

	// 1. Normal playback before sync
	normalTime := time.Unix(5, 0).UTC()
	normalState := playbackSvc.ComputePlaybackState(context.Background(), window, originalPlaylist, normalTime)
	if normalState.IsSyncOverride {
		t.Errorf("expected IsSyncOverride=false before sync")
	}
	if normalState.CurrentPlaylistItem.MediaID != 1 {
		t.Errorf("expected normal media 1, got %d", normalState.CurrentPlaylistItem.MediaID)
	}

	// 2. Start sync
	event, err := syncSvc.StartSync(context.Background(), models.StartSyncRequest{
		MediaID:  2,
		Duration: 15,
	})
	if err != nil {
		t.Fatalf("failed to start sync: %v", err)
	}

	// 3. Playback during sync window
	syncMidTime := event.StartAt.Add(5 * time.Second)
	syncState := playbackSvc.ComputePlaybackState(context.Background(), window, originalPlaylist, syncMidTime)
	if !syncState.IsSyncOverride {
		t.Errorf("expected IsSyncOverride=true during sync window")
	}
	if syncState.ActiveSyncEvent.MediaID != 2 {
		t.Errorf("expected active sync media 2, got %d", syncState.ActiveSyncEvent.MediaID)
	}

	// 4. Playback after sync expires (original playlist must remain intact!)
	postSyncTime := event.EndAt.Add(1 * time.Second)
	_ = syncRepo.ExpirePastEvents(context.Background(), postSyncTime)
	postSyncState := playbackSvc.ComputePlaybackState(context.Background(), window, originalPlaylist, postSyncTime)

	if postSyncState.IsSyncOverride {
		t.Errorf("expected IsSyncOverride=false after sync window ends")
	}
	// Verify original playlist items were not mutated
	if len(originalPlaylist) != 2 || originalPlaylist[0].MediaID != 1 || originalPlaylist[1].MediaID != 3 {
		t.Errorf("original playlist was corrupted by sync operation!")
	}
}
