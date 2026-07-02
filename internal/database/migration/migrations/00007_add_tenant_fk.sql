-- +goose Up
ALTER TABLE users ADD CONSTRAINT fk_user_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE dead_letters ADD CONSTRAINT fk_dl_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- +goose Down
ALTER TABLE users DROP FOREIGN KEY fk_user_tenant;
ALTER TABLE dead_letters DROP FOREIGN KEY fk_dl_tenant;
