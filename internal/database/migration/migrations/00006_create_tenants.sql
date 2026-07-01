-- +goose Up
CREATE TABLE tenants (
    id BIGINT UNSIGNED PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(64) NOT NULL,
    active TINYINT(1) NOT NULL DEFAULT 1,
    UNIQUE INDEX idx_tenant_code (code),
    INDEX idx_tenant_active (active),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS tenants;
