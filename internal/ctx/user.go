package ctx

import "context"

type userUniCodeKey struct{}

func WithUserUniCode(ctx context.Context, uniCode string) context.Context {
	return context.WithValue(ctx, userUniCodeKey{}, uniCode)
}

func GetUserUniCodeFromContext(ctx context.Context) string {
	value, _ := ctx.Value(userUniCodeKey{}).(string)
	return value
}
