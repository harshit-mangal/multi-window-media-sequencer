package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"media-sequencer-backend/internal/models"
)

type SyncRepository interface {
	Create(ctx context.Context, event models.SyncEvent) (*models.SyncEvent, error)
	GetCurrentActive(ctx context.Context, now time.Time) (*models.SyncEvent, error)
	GetByID(ctx context.Context, id string) (*models.SyncEvent, error)
	ExpirePastEvents(ctx context.Context, now time.Time) error
}

type pgSyncRepository struct {
	pool *pgxpool.Pool
}

func NewSyncRepository(pool *pgxpool.Pool) SyncRepository {
	return &pgSyncRepository{pool: pool}
}

func (r *pgSyncRepository) Create(ctx context.Context, event models.SyncEvent) (*models.SyncEvent, error) {
	// First mark any currently active syncs as superseded to maintain clean state
	_, _ = r.pool.Exec(ctx, `UPDATE sync_events SET status = 'superseded' WHERE status = 'active' AND end_at > $1`, event.StartAt)

	query := `
		INSERT INTO sync_events (id, media_id, duration, start_at, end_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, media_id, duration, start_at, end_at, status, created_at
	`
	var res models.SyncEvent
	err := r.pool.QueryRow(ctx, query, event.ID, event.MediaID, event.Duration, event.StartAt, event.EndAt, event.Status).Scan(
		&res.ID, &res.MediaID, &res.Duration, &res.StartAt, &res.EndAt, &res.Status, &res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync event: %w", err)
	}
	return r.GetByID(ctx, res.ID)
}

func (r *pgSyncRepository) GetCurrentActive(ctx context.Context, now time.Time) (*models.SyncEvent, error) {
	// Look for event where status = 'active' AND end_at > now
	query := `
		SELECT 
			s.id, s.media_id, s.duration, s.start_at, s.end_at, s.status, s.created_at,
			m.id, m.name, m.type, m.url, m.duration, m.created_at, m.updated_at
		FROM sync_events s
		INNER JOIN media m ON s.media_id = m.id
		WHERE s.status = 'active' AND s.end_at > $1
		ORDER BY s.start_at DESC
		LIMIT 1
	`
	var event models.SyncEvent
	var m models.Media
	err := r.pool.QueryRow(ctx, query, now).Scan(
		&event.ID, &event.MediaID, &event.Duration, &event.StartAt, &event.EndAt, &event.Status, &event.CreatedAt,
		&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query active sync event: %w", err)
	}
	event.Media = &m
	return &event, nil
}

func (r *pgSyncRepository) GetByID(ctx context.Context, id string) (*models.SyncEvent, error) {
	query := `
		SELECT 
			s.id, s.media_id, s.duration, s.start_at, s.end_at, s.status, s.created_at,
			m.id, m.name, m.type, m.url, m.duration, m.created_at, m.updated_at
		FROM sync_events s
		INNER JOIN media m ON s.media_id = m.id
		WHERE s.id = $1
	`
	var event models.SyncEvent
	var m models.Media
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.MediaID, &event.Duration, &event.StartAt, &event.EndAt, &event.Status, &event.CreatedAt,
		&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query sync event by ID: %w", err)
	}
	event.Media = &m
	return &event, nil
}

func (r *pgSyncRepository) ExpirePastEvents(ctx context.Context, now time.Time) error {
	query := `UPDATE sync_events SET status = 'completed' WHERE status = 'active' AND end_at <= $1`
	_, err := r.pool.Exec(ctx, query, now)
	return err
}
