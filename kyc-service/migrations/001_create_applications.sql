-- +goose Up
CREATE TABLE IF NOT EXISTS applications
(
    id               UUID PRIMARY KEY     DEFAULT uuidv7(),
    user_id          UUID        NOT NULL,
    status           TEXT        NOT NULL DEFAULT 'DRAFT',
    claimed_by       UUID,
    claimed_at       TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_applications_active_user
    ON applications (user_id)
    WHERE status NOT IN ('APPROVED', 'REJECTED', 'AUTO_REJECTED');

-- +goose Down
DROP TABLE IF EXISTS applications;