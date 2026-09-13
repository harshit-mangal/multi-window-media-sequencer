package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/handlers"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
	"media-sequencer-backend/internal/websocket"
)

type mockWindowRepo struct {
	windows map[int]models.Window
}

func (m *mockWindowRepo) GetAll(ctx context.Context) ([]models.Window, error) {
	var list []models.Window
	for _, w := range m.windows {
		list = append(list, w)
	}
	return list, nil
}

func (m *mockWindowRepo) GetByID(ctx context.Context, id int) (*models.Window, error) {
	if w, ok := m.windows[id]; ok {
		return &w, nil
	}
	return nil, nil
}

func (m *mockWindowRepo) Create(ctx context.Context, name string) (*models.Window, error) {
	w := models.Window{ID: len(m.windows) + 1, Name: name}
	m.windows[w.ID] = w
	return &w, nil
}

type mockPlaylistRepo struct {
	items map[int][]models.PlaylistItem
}

func (m *mockPlaylistRepo) GetByWindowID(ctx context.Context, windowID int) ([]models.PlaylistItem, error) {
	return m.items[windowID], nil
}

func (m *mockPlaylistRepo) GetItemByID(ctx context.Context, itemID int) (*models.PlaylistItem, error) {
	for _, list := range m.items {
		for _, item := range list {
			if item.ID == itemID {
				return &item, nil
			}
		}
	}
	return nil, nil
}

func (m *mockPlaylistRepo) AddItem(ctx context.Context, windowID int, mediaID int, targetPos *int) (*models.PlaylistItem, error) {
	item := models.PlaylistItem{
		ID:       100 + len(m.items[windowID]),
		WindowID: windowID,
		MediaID:  mediaID,
		Position: len(m.items[windowID]),
	}
	m.items[windowID] = append(m.items[windowID], item)
	return &item, nil
}

func (m *mockPlaylistRepo) UpdateItem(ctx context.Context, itemID int, targetPos *int, mediaID *int) (*models.PlaylistItem, error) {
	for wID, list := range m.items {
		for i, item := range list {
			if item.ID == itemID {
				if mediaID != nil {
					m.items[wID][i].MediaID = *mediaID
				}
				return &m.items[wID][i], nil
			}
		}
	}
	return nil, nil
}

func (m *mockPlaylistRepo) RemoveItem(ctx context.Context, windowID int, itemID int) error {
	list := m.items[windowID]
	var newList []models.PlaylistItem
	for _, item := range list {
		if item.ID != itemID {
			newList = append(newList, item)
		}
	}
	m.items[windowID] = newList
	return nil
}

func setupTestRouter() *handlers.RouterDependencies {
	cfg := &config.Config{
		Port:             "8080",
		AllowedOrigins:   "*",
		SyncBufferSec:    2,
		CycleDurationSec: 18000,
	}

	windowRepo := &mockWindowRepo{
		windows: map[int]models.Window{
			1: {ID: 1, Name: "Window 1"},
			2: {ID: 2, Name: "Window 2"},
		},
	}
	mediaRepo := newMockMediaRepo()
	playlistRepo := &mockPlaylistRepo{
		items: make(map[int][]models.PlaylistItem),
	}
	syncRepo := newMockSyncRepo()
	wsHub := websocket.NewHub()

	playbackSvc := services.NewPlaybackService(cfg, syncRepo)
	syncSvc := services.NewSyncService(cfg, syncRepo, mediaRepo, wsHub)
	windowSvc := services.NewWindowService(windowRepo, playlistRepo, playbackSvc)
	mediaSvc := services.NewMediaService(mediaRepo)
	playlistSvc := services.NewPlaylistService(playlistRepo, windowRepo, mediaRepo, wsHub)

	return &handlers.RouterDependencies{
		Config:          cfg,
		WindowHandler:   handlers.NewWindowHandler(windowSvc),
		MediaHandler:    handlers.NewMediaHandler(mediaSvc),
		PlaylistHandler: handlers.NewPlaylistHandler(playlistSvc),
		SyncHandler:     handlers.NewSyncHandler(syncSvc),
		WSHandler:       handlers.NewWebSocketHandler(wsHub),
	}
}

func TestAPI_GetWindows(t *testing.T) {
	deps := setupTestRouter()
	router := handlers.SetupRouter(*deps)

	req, _ := http.NewRequest(http.MethodGet, "/api/windows", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp["data"].([]any)
	if !ok || len(data) != 2 {
		t.Fatalf("expected 2 windows in response, got %v", resp["data"])
	}
}

func TestAPI_StartSyncEndpoint(t *testing.T) {
	deps := setupTestRouter()
	router := handlers.SetupRouter(*deps)

	payload := models.StartSyncRequest{
		MediaID:  2,
		Duration: 15,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify current sync
	getReq, _ := http.NewRequest(http.MethodGet, "/api/sync/current", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /api/sync/current, got %d", getRec.Code)
	}
}

func TestAPI_PlaylistOperations(t *testing.T) {
	deps := setupTestRouter()
	router := handlers.SetupRouter(*deps)

	// 1. Add item to Window 1
	addPayload := models.AddToPlaylistRequest{MediaID: 2}
	addBody, _ := json.Marshal(addPayload)
	addReq, _ := http.NewRequest(http.MethodPost, "/api/windows/1/playlist", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addRec := httptest.NewRecorder()
	router.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on adding playlist item, got %d: %s", addRec.Code, addRec.Body.String())
	}

	// 2. Fetch playlist
	getReq, _ := http.NewRequest(http.MethodGet, "/api/windows/1/playlist", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on getting playlist, got %d", getRec.Code)
	}

	// 3. Delete item
	delReq, _ := http.NewRequest(http.MethodDelete, "/api/windows/1/playlist/100", nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on deleting item, got %d", delRec.Code)
	}
}

func TestAPI_MediaCRUD(t *testing.T) {
	deps := setupTestRouter()
	router := handlers.SetupRouter(*deps)

	// Create Media
	createPayload := models.CreateMediaRequest{
		Name:     "Test Image",
		Type:     models.MediaTypeImage,
		URL:      "https://example.com/test.jpg",
		Duration: 12,
	}
	body, _ := json.Marshal(createPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on media creation, got %d: %s", rec.Code, rec.Body.String())
	}

	// Invalid Media Type
	invalidPayload := models.CreateMediaRequest{
		Name:     "Invalid Type Item",
		Type:     "audio",
		URL:      "https://example.com/test.mp3",
		Duration: 10,
	}
	invBody, _ := json.Marshal(invalidPayload)
	invReq, _ := http.NewRequest(http.MethodPost, "/api/media", bytes.NewBuffer(invBody))
	invReq.Header.Set("Content-Type", "application/json")
	invRec := httptest.NewRecorder()
	router.ServeHTTP(invRec, invReq)

	if invRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on invalid media type, got %d", invRec.Code)
	}
}
