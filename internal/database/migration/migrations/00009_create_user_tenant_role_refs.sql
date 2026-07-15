-- +goose Up
CREATE TABLE user_tenant_role_refs (
    user_id BIGINT UNSIGNED NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, tenant_id, role_id),
    INDEX idx_utrr_tenant_role (tenant_id, role_id),
    CONSTRAINT fk_utrr_membership FOREIGN KEY (user_id, tenant_id)
        REFERENCES user_tenant_refs(user_id, tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_utrr_role FOREIGN KEY (role_id) REFERENCES user_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Ensure every existing user's primary tenant is represented as a membership.
INSERT IGNORE INTO user_tenant_refs (user_id, tenant_id)
SELECT id, tenant_id FROM users;

-- Legacy roles had no tenant scope. Preserve them only in each user's primary tenant.
INSERT IGNORE INTO user_tenant_role_refs (user_id, tenant_id, role_id)
SELECT u.id, u.tenant_id, urr.user_role_id
FROM users u
JOIN user_role_refs urr ON urr.user_id = u.id;

DROP TABLE user_role_refs;

-- +goose Down
CREATE TABLE user_role_refs (
    user_id BIGINT UNSIGNED NOT NULL,
    user_role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, user_role_id),
    CONSTRAINT fk_user_role_refs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_refs_role FOREIGN KEY (user_role_id) REFERENCES user_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO user_role_refs (user_id, user_role_id)
SELECT user_id, role_id FROM user_tenant_role_refs;

DROP TABLE user_tenant_role_refs;
