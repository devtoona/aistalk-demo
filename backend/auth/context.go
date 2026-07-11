package auth

import "context"

type ctxKey int

const uidKey ctxKey = 1

func WithUID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, uidKey, uid)
}

func UIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(uidKey).(string)
	if !ok || uid == "" {
		return "", false
	}
	return uid, true
}
