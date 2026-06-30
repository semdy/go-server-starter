-- +goose Up
CREATE TABLE user_roles (
    id BIGINT UNSIGNED PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    code VARCHAR(50) NOT NULL,
    enabled TINYINT(1) DEFAULT 1,
    UNIQUE INDEX idx_code (code),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS user_roles;
