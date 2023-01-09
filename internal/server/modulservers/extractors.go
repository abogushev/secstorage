package modulservers

import (
	"context"
	"secstorage/internal/api"
)

func extractUserId(ctx context.Context) api.UserId {
	return ctx.Value("userId").(api.UserId)
}
