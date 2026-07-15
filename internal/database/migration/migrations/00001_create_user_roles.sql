-- +goose Up
CREATE TABLE user_roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    built_in TINYINT(1) NOT NULL DEFAULT 0,
    enabled TINYINT(1) DEFAULT 1,
    UNIQUE INDEX idx_tenant_role_code (tenant_id, code),
    INDEX idx_user_roles_tenant_id (tenant_id),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS user_roles;
