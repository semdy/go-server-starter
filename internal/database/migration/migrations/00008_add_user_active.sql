-- +goose Up
ALTER TABLE users ADD COLUMN active TINYINT(1) NOT NULL DEFAULT 1 AFTER uni_code;

-- +goose Down
ALTER TABLE users DROP COLUMN active;
