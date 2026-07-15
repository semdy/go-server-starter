-- +goose Up
CREATE TABLE user_tenant_role_refs (
    user_id BIGINT UNSIGNED NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, tenant_id, role_id),
    INDEX idx_utrr_tenant_role (tenant_id, role_id),
    CONSTRAINT fk_utrr_membership FOREIGN KEY (user_id, tenant_id) REFERENCES user_tenant_refs(user_id, tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_utrr_role FOREIGN KEY (role_id) REFERENCES user_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS user_tenant_role_refs;
