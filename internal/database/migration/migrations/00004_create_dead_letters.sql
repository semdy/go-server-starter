-- +goose Up
CREATE TABLE dead_letters (
    id BIGINT UNSIGNED PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    task_type VARCHAR(100) NOT NULL,
    task_id VARCHAR(255) NOT NULL,
    queue VARCHAR(50) NOT NULL DEFAULT 'default',
    payload LONGBLOB,
    error TEXT,
    attempt INT DEFAULT 0,
    max_retry INT DEFAULT 0,
    failed_at DATETIME(3) NOT NULL,
    is_retried TINYINT(1) DEFAULT 0,
    retried_at DATETIME(3) NULL,
    UNIQUE INDEX idx_task_id (task_id),
    INDEX idx_task_type (task_type),
    INDEX idx_queue (queue),
    INDEX idx_failed_at (failed_at),
    INDEX idx_is_retried (is_retried),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS dead_letters;
