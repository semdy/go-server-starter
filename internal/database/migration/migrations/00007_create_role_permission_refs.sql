-- +goose Up
CREATE TABLE role_permission_refs (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_rpr_role FOREIGN KEY (role_id) REFERENCES user_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_rpr_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS role_permission_refs;
