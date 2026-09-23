CREATE TABLE IF NOT EXISTS saved_routes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    vehicle_count INTEGER NOT NULL DEFAULT 1,
    total_distance_km DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_duration_minutes DOUBLE PRECISION NOT NULL DEFAULT 0,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_saved_routes_created_at ON saved_routes (created_at DESC);