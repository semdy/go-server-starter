-- +goose Up
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    uni_code VARCHAR(64) NOT NULL,
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
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS users;
