-- +goose Up
ALTER TABLE user_roles
    DROP INDEX idx_code,
    ADD COLUMN tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id,
    ADD COLUMN name VARCHAR(100) NOT NULL DEFAULT '' AFTER code,
    ADD COLUMN description VARCHAR(255) NOT NULL DEFAULT '' AFTER name,
    ADD COLUMN built_in TINYINT(1) NOT NULL DEFAULT 0 AFTER description,
    ADD UNIQUE INDEX idx_tenant_role_code (tenant_id, code),
    ADD INDEX idx_user_roles_tenant_id (tenant_id);

UPDATE user_roles
SET name = code, built_in = 1, tenant_id = 0
WHERE code IN ('super_admin', 'admin', 'guest', 'user', 'user_vip', 'user_svip');

-- +goose Down
ALTER TABLE user_roles
    DROP INDEX idx_tenant_role_code,
    DROP INDEX idx_user_roles_tenant_id,
    DROP COLUMN tenant_id,
    DROP COLUMN name,
    DROP COLUMN description,
    DROP COLUMN built_in,
    ADD UNIQUE INDEX idx_code (code);
