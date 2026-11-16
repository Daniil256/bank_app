-- +goose Up
CREATE TABLE IF NOT EXISTS sanctions_list
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    birth_date DATE NOT NULL,
    reason     TEXT
);

INSERT INTO sanctions_list (first_name, last_name, birth_date, reason)
VALUES ('Ivan', 'Ivanov', '1990-01-01', 'sanctions match');

-- +goose Down
DROP TABLE IF EXISTS sanctions_list CASCADE;