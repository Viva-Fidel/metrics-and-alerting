CREATE TABLE IF NOT EXISTS metrics (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('gauge', 'counter')),
    delta BIGINT,
    value DOUBLE PRECISION,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (type = 'gauge' AND value IS NOT NULL AND delta IS NULL) OR
        (type = 'counter' AND delta IS NOT NULL AND value IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS metrics_type_idx ON metrics (type);
