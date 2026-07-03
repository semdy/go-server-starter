-- +goose Up
CREATE TABLE users (
    id BIGINT UNSIGNED PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    uni_code VARCHAR(64) NOT NULL,
    active TINYINT(1) NOT NULL DEFAULT 1,
    tenant_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(255),
    mobile VARCHAR(32),
    country_code VARCHAR(8),
    `desc` TEXT,
    password VARCHAR(255),
    salt VARCHAR(255),
    nickname VARCHAR(255),
    avatar_url VARCHAR(512),
    UNIQUE INDEX idx_uni_code (uni_code),
    INDEX idx_email (email),
    INDEX idx_mobile_country (mobile, country_code),
    INDEX idx_nickname (nickname),
    UNIQUE INDEX idx_email_tenant (email, tenant_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_deleted_at (deleted_at),
    CONSTRAINT fk_user_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS users;
