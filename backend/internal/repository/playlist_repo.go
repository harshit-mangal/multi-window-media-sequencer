package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"media-sequencer-backend/internal/models"
)

type PlaylistRepository interface {
	GetByWindowID(ctx context.Context, windowID int) ([]models.PlaylistItem, error)
	GetItemByID(ctx context.Context, itemID int) (*models.PlaylistItem, error)
	AddItem(ctx context.Context, windowID int, mediaID int, targetPos *int) (*models.PlaylistItem, error)
	UpdateItem(ctx context.Context, itemID int, targetPos *int, mediaID *int) (*models.PlaylistItem, error)
	RemoveItem(ctx context.Context, windowID int, itemID int) error
}

type pgPlaylistRepository struct {
	pool *pgxpool.Pool
}

func NewPlaylistRepository(pool *pgxpool.Pool) PlaylistRepository {
	return &pgPlaylistRepository{pool: pool}
}

func (r *pgPlaylistRepository) GetByWindowID(ctx context.Context, windowID int) ([]models.PlaylistItem, error) {
	query := `
		SELECT 
			p.id, p.window_id, p.media_id, p.position, p.created_at, p.updated_at,
			m.id, m.name, m.type, m.url, m.duration, m.created_at, m.updated_at
		FROM playlist_items p
		INNER JOIN media m ON p.media_id = m.id
		WHERE p.window_id = $1
		ORDER BY p.position ASC
	`
	rows, err := r.pool.Query(ctx, query, windowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query playlist items: %w", err)
	}
	defer rows.Close()

	var items []models.PlaylistItem
	for rows.Next() {
		var item models.PlaylistItem
		var m models.Media
		err := rows.Scan(
			&item.ID, &item.WindowID, &item.MediaID, &item.Position, &item.CreatedAt, &item.UpdatedAt,
			&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan playlist item: %w", err)
		}
		item.Media = &m
		items = append(items, item)
	}
	return items, nil
}

func (r *pgPlaylistRepository) GetItemByID(ctx context.Context, itemID int) (*models.PlaylistItem, error) {
	query := `
		SELECT 
			p.id, p.window_id, p.media_id, p.position, p.created_at, p.updated_at,
			m.id, m.name, m.type, m.url, m.duration, m.created_at, m.updated_at
		FROM playlist_items p
		INNER JOIN media m ON p.media_id = m.id
		WHERE p.id = $1
	`
	var item models.PlaylistItem
	var m models.Media
	err := r.pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID, &item.WindowID, &item.MediaID, &item.Position, &item.CreatedAt, &item.UpdatedAt,
		&m.ID, &m.Name, &m.Type, &m.URL, &m.Duration, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query playlist item by ID: %w", err)
	}
	item.Media = &m
	return &item, nil
}

func (r *pgPlaylistRepository) AddItem(ctx context.Context, windowID int, mediaID int, targetPos *int) (*models.PlaylistItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Determine max position
	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM playlist_items WHERE window_id = $1", windowID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count playlist items: %w", err)
	}

	insertPos := count
	if targetPos != nil && *targetPos >= 0 && *targetPos < count {
		insertPos = *targetPos
		// Shift existing items at or after insertPos to right
		// To avoid unique constraint collisions during shift, update in reverse order
		shiftQuery := `
			UPDATE playlist_items 
			SET position = position + 1, updated_at = CURRENT_TIMESTAMP
			WHERE window_id = $1 AND position >= $2
		`
		_, err = tx.Exec(ctx, shiftQuery, windowID, insertPos)
		if err != nil {
			return nil, fmt.Errorf("failed to shift playlist item positions: %w", err)
		}
	}

	insertQuery := `
		INSERT INTO playlist_items (window_id, media_id, position)
		VALUES ($1, $2, $3)
		RETURNING id, window_id, media_id, position, created_at, updated_at
	`
	var item models.PlaylistItem
	err = tx.QueryRow(ctx, insertQuery, windowID, mediaID, insertPos).Scan(
		&item.ID, &item.WindowID, &item.MediaID, &item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert playlist item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetItemByID(ctx, item.ID)
}

func (r *pgPlaylistRepository) UpdateItem(ctx context.Context, itemID int, targetPos *int, mediaID *int) (*models.PlaylistItem, error) {
	current, err := r.GetItemByID(ctx, itemID)
	if err != nil || current == nil {
		return nil, fmt.Errorf("playlist item not found: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	newMediaID := current.MediaID
	if mediaID != nil && *mediaID > 0 {
		newMediaID = *mediaID
	}

	if targetPos != nil && *targetPos != current.Position {
		newPos := *targetPos
		// Temporarily set current item position to -1 to avoid unique constraint collision
		_, err = tx.Exec(ctx, "UPDATE playlist_items SET position = -1 WHERE id = $1", itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to temp shift item: %w", err)
		}

		if newPos > current.Position {
			// Shifting right: items between (current.Position, newPos] shift left (-1)
			_, err = tx.Exec(ctx, `
				UPDATE playlist_items 
				SET position = position - 1 
				WHERE window_id = $1 AND position > $2 AND position <= $3
			`, current.WindowID, current.Position, newPos)
		} else {
			// Shifting left: items between [newPos, current.Position) shift right (+1)
			_, err = tx.Exec(ctx, `
				UPDATE playlist_items 
				SET position = position + 1 
				WHERE window_id = $1 AND position >= $2 AND position < $3
			`, current.WindowID, newPos, current.Position)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to reorder items: %w", err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE playlist_items 
			SET position = $1, media_id = $2, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $3
		`, newPos, newMediaID, itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to update item position: %w", err)
		}
	} else {
		// Only updating media
		_, err = tx.Exec(ctx, `
			UPDATE playlist_items 
			SET media_id = $1, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $2
		`, newMediaID, itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to update item media: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetItemByID(ctx, itemID)
}

func (r *pgPlaylistRepository) RemoveItem(ctx context.Context, windowID int, itemID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var removedPos int
	err = tx.QueryRow(ctx, "DELETE FROM playlist_items WHERE id = $1 AND window_id = $2 RETURNING position", itemID, windowID).Scan(&removedPos)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("playlist item not found")
		}
		return fmt.Errorf("failed to delete playlist item: %w", err)
	}

	// Compact remaining positions
	_, err = tx.Exec(ctx, `
		UPDATE playlist_items 
		SET position = position - 1, updated_at = CURRENT_TIMESTAMP
		WHERE window_id = $1 AND position > $2
	`, windowID, removedPos)
	if err != nil {
		return fmt.Errorf("failed to compact playlist positions: %w", err)
	}

	return tx.Commit(ctx)
}
