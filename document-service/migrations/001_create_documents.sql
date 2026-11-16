-- +goose Up
CREATE TABLE IF NOT EXISTS documents
(
    id           UUID PRIMARY KEY     DEFAULT uuidv7(),
    customer_id  UUID        NOT NULL,
    doc_type     TEXT        NOT NULL,
    object_key   TEXT        NOT NULL UNIQUE,
    status       TEXT        NOT NULL DEFAULT 'UPLOADED',
    size_bytes   BIGINT,
    mime_type    TEXT,
    sha256       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_documents_customer ON documents (customer_id);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents (doc_type);

-- +goose Down
DROP TABLE IF EXISTS documents;