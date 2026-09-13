-- init_schema.up.sql

CREATE TABLE IF NOT EXISTS windows (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS media (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('image', 'video', 'blank')),
    url TEXT NOT NULL,
    duration INT NOT NULL CHECK (duration > 0), -- duration in seconds
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS playlist_items (
    id SERIAL PRIMARY KEY,
    window_id INT NOT NULL REFERENCES windows(id) ON DELETE CASCADE,
    media_id INT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    position INT NOT NULL CHECK (position >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_window_position UNIQUE (window_id, position)
);

CREATE TABLE IF NOT EXISTS sync_events (
    id VARCHAR(64) PRIMARY KEY,
    media_id INT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    duration INT NOT NULL CHECK (duration > 0),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_playlist_items_window ON playlist_items(window_id, position);
CREATE INDEX IF NOT EXISTS idx_sync_events_status_time ON sync_events(status, start_at, end_at);
