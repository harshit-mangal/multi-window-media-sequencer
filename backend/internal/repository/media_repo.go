package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"media-sequencer-backend/internal/models"
)

type MediaRepository interface {
	GetAll(ctx context.Context) ([]models.Media, error)
	GetByID(ctx context.Context, id int) (*models.Media, error)
	Create(ctx context.Context, req models.CreateMediaRequest) (*models.Media, error)
	Delete(ctx context.Context, id int) error
}

type pgMediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return &pgMediaRepository{pool: pool}
}

func (r *pgMediaRepository) GetAll(ctx context.Context) ([]models.Media, error) {
	query := `
		SELECT id, name, type, url, duration, created_at, updated_at
		FROM media
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query media: %w", err)
	}
	defer rows.Close()

	var mediaList []models.Media
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		mediaList = append(mediaList, m)
	}
	return mediaList, nil
}

func (r *pgMediaRepository) GetByID(ctx context.Context, id int) (*models.Media, error) {
	query := `
		SELECT id, name, type, url, duration, created_at, updated_at
		FROM media
		WHERE id = $1
	`
	var m models.Media
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query media by ID: %w", err)
	}
	return &m, nil
}

func (r *pgMediaRepository) Create(ctx context.Context, req models.CreateMediaRequest) (*models.Media, error) {
	query := `
		INSERT INTO media (name, type, url, duration)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, type, url, duration, created_at, updated_at
	`
	var m models.Media
	err := r.pool.QueryRow(ctx, query, req.Name, req.Type, req.URL, req.Duration).
		Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}
	return &m, nil
}

func (r *pgMediaRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM media WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("media not found")
	}
	return nil
}
