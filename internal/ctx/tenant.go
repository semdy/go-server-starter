package ctx

import "context"

type tenantKey struct{}

// WithTenant stores tenant_id in a context.Context.
func WithTenant(ctx context.Context, tenantID uint64) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// GetTenantID extracts tenant_id from a context.Context. Returns 0 if not set.
func GetTenantID(ctx context.Context) uint64 {
	v, _ := ctx.Value(tenantKey{}).(uint64)
	return v
}
