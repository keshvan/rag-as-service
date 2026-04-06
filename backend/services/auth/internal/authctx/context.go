package authctx

import "context"

type contextKey string

const (
	UserGUIDKey  contextKey = "user_guid"
	SessionIDKey contextKey = "session_id"
)

func WithUserGUID(ctx context.Context, guid string) context.Context {
	return context.WithValue(ctx, UserGUIDKey, guid)
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

func UserGUIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(UserGUIDKey).(string)
	return v, ok
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(SessionIDKey).(string)
	return v, ok
}
