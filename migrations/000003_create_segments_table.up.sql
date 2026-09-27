CREATE TABLE IF NOT EXISTS segments (
    id TEXT PRIMARY KEY,
    system_id TEXT NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    shape TEXT NOT NULL,
    width INTEGER,
    height INTEGER,
    diameter INTEGER,
    length DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);