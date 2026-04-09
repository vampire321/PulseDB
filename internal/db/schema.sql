CREATE TABLE IF NOT EXISTS monitors (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    url         TEXT        NOT NULL,
    interval_s  INTEGER     NOT NULL DEFAULT 60,
    status      TEXT        NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_monitors_active
    ON monitors(status) WHERE status = 'active';
    