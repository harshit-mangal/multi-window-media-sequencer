package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"media-sequencer-backend/internal/models"
)

type WindowRepository interface {
	GetAll(ctx context.Context) ([]models.Window, error)
	GetByID(ctx context.Context, id int) (*models.Window, error)
	Create(ctx context.Context, name string) (*models.Window, error)
}

type pgWindowRepository struct {
	pool *pgxpool.Pool
}

func NewWindowRepository(pool *pgxpool.Pool) WindowRepository {
	return &pgWindowRepository{pool: pool}
}

func (r *pgWindowRepository) GetAll(ctx context.Context) ([]models.Window, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM windows
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query windows: %w", err)
	}
	defer rows.Close()

	var windows []models.Window
	for rows.Next() {
		var w models.Window
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan window: %w", err)
		}
		windows = append(windows, w)
	}
	return windows, nil
}

func (r *pgWindowRepository) GetByID(ctx context.Context, id int) (*models.Window, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM windows
		WHERE id = $1
	`
	var w models.Window
	err := r.pool.QueryRow(ctx, query, id).Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query window by ID: %w", err)
	}
	return &w, nil
}

func (r *pgWindowRepository) Create(ctx context.Context, name string) (*models.Window, error) {
	query := `
		INSERT INTO windows (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`
	var w models.Window
	err := r.pool.QueryRow(ctx, query, name).Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	return &w, nil
}
