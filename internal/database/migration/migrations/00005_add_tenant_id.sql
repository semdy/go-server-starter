-- +goose Up
ALTER TABLE users ADD COLUMN tenant_id VARCHAR(64) NOT NULL DEFAULT 'default' AFTER uni_code;
ALTER TABLE dead_letters ADD COLUMN tenant_id VARCHAR(64) NOT NULL DEFAULT 'default' AFTER id;
CREATE INDEX idx_tenant_id ON users (tenant_id);
CREATE INDEX idx_dl_tenant_id ON dead_letters (tenant_id);

-- +goose Down
ALTER TABLE users DROP INDEX idx_tenant_id;
ALTER TABLE users DROP COLUMN tenant_id;
ALTER TABLE dead_letters DROP INDEX idx_dl_tenant_id;
ALTER TABLE dead_letters DROP COLUMN tenant_id;
