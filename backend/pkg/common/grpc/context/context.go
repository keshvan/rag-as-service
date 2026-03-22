package context

import (
	"context"
)

type ctxKey string

const OrgIDKey ctxKey = "organization_id"

// FromContext извлекает ID организации из контекста
func FromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(OrgIDKey).(string)
	return orgID, ok
}

// ToContext кладет ID организации в контекст
func ToContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDKey, orgID)
}
