package context

import (
	"context"
)

type ctxKey string

const OrgIDKey ctxKey = "organization_id"

func FromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(OrgIDKey).(string)
	return orgID, ok
}

func ToContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDKey, orgID)
}
