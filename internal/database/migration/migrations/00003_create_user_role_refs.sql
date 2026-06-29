-- +goose Up
CREATE TABLE user_role_refs (
    user_id BIGINT UNSIGNED NOT NULL,
    user_role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, user_role_id),
    CONSTRAINT fk_user_role_refs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_refs_role FOREIGN KEY (user_role_id) REFERENCES user_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS user_role_refs;
