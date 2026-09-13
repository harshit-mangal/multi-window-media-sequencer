package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
	"media-sequencer-backend/internal/websocket"
)

type SyncService interface {
	StartSync(ctx context.Context, req models.StartSyncRequest) (*models.SyncEvent, error)
	GetCurrentSync(ctx context.Context) (*models.SyncCurrentResponse, error)
}

type syncService struct {
	cfg       *config.Config
	syncRepo  repository.SyncRepository
	mediaRepo repository.MediaRepository
	wsHub     *websocket.Hub
	activeMu  sync.Mutex
	endTimer  *time.Timer
}

func NewSyncService(
	cfg *config.Config,
	syncRepo repository.SyncRepository,
	mediaRepo repository.MediaRepository,
	wsHub *websocket.Hub,
) SyncService {
	return &syncService{
		cfg:       cfg,
		syncRepo:  syncRepo,
		mediaRepo: mediaRepo,
		wsHub:     wsHub,
	}
}

func (s *syncService) StartSync(ctx context.Context, req models.StartSyncRequest) (*models.SyncEvent, error) {
	s.activeMu.Lock()
	defer s.activeMu.Unlock()

	// 1. Validate target media exists
	media, err := s.mediaRepo.GetByID(ctx, req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve media: %w", err)
	}
	if media == nil {
		return nil, errors.New("media item not found")
	}

	// 2. Validate duration
	duration := req.Duration
	if duration <= 0 {
		duration = media.Duration // Default to media's own duration if not specified
	}
	if duration <= 0 {
		duration = 15 // Fallback default
	}

	// 3. Generate unique sync identifier
	syncID := uuid.New().String()

	// 4. Calculate shared synchronized start and end timestamps
	now := time.Now().UTC()
	startAt := now.Add(time.Duration(s.cfg.SyncBufferSec) * time.Second)
	endAt := startAt.Add(time.Duration(duration) * time.Second)

	// 5. Store sync event in persistent PostgreSQL database
	eventToCreate := models.SyncEvent{
		ID:        syncID,
		MediaID:   media.ID,
		Duration:  duration,
		StartAt:   startAt,
		EndAt:     endAt,
		Status:    "active",
		Media:     media,
		CreatedAt: now,
	}

	createdEvent, err := s.syncRepo.Create(ctx, eventToCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to persist sync event: %w", err)
	}

	// 6. Broadcast SYNC_START event through WebSocket hub to all connected window clients
	if s.wsHub != nil {
		s.wsHub.BroadcastSyncStart(syncID, media.ID, duration, startAt, endAt, media)
	}

	// 7. Schedule automatic SYNC_END broadcast and status expiration
	if s.endTimer != nil {
		s.endTimer.Stop()
	}

	timeUntilEnd := time.Until(endAt)
	s.endTimer = time.AfterFunc(timeUntilEnd, func() {
		bgCtx := context.Background()
		if err := s.syncRepo.ExpirePastEvents(bgCtx, time.Now().UTC()); err != nil {
			log.Printf("[SYNC_ERROR] Failed to mark sync event expired: %v", err)
		}
		if s.wsHub != nil {
			s.wsHub.BroadcastSyncEnd(syncID)
		}
	})

	return createdEvent, nil
}

func (s *syncService) GetCurrentSync(ctx context.Context) (*models.SyncCurrentResponse, error) {
	now := time.Now().UTC()

	// Clean up past events first
	_ = s.syncRepo.ExpirePastEvents(ctx, now)

	activeEvent, err := s.syncRepo.GetCurrentActive(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active sync event: %w", err)
	}

	if activeEvent == nil {
		return &models.SyncCurrentResponse{
			Active:     false,
			ServerTime: now,
		}, nil
	}

	remainingSec := int(activeEvent.EndAt.Sub(now).Seconds())
	if remainingSec < 0 {
		remainingSec = 0
	}

	return &models.SyncCurrentResponse{
		Active:       true,
		SyncEvent:    activeEvent,
		RemainingSec: remainingSec,
		ServerTime:   now,
	}, nil
}
