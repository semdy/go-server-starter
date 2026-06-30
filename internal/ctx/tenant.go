package ctx

import "context"

type tenantKey struct{}

// WithTenant stores tenant_id in a context.Context.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// GetTenantID extracts tenant_id from a context.Context.
// Returns empty string if not set.
func GetTenantID(ctx context.Context) string {
	if v, ok := ctx.Value(tenantKey{}).(string); ok {
		return v
	}
	return ""
}
