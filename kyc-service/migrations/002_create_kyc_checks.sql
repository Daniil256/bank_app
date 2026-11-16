-- +goose Up
CREATE TABLE IF NOT EXISTS kyc_checks
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    application_id UUID        NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    check_type     TEXT        NOT NULL,
    result         TEXT        NOT NULL,
    details        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_kyc_checks_application ON kyc_checks (application_id);

-- +goose Down
DROP TABLE IF EXISTS kyc_checks;