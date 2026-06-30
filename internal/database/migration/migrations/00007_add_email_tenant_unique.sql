-- +goose Up
ALTER TABLE users ADD UNIQUE INDEX idx_email_tenant (email, tenant_id);

-- +goose Down
ALTER TABLE users DROP INDEX idx_email_tenant;
