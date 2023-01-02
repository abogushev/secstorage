package implementations

import (
	"context"
	"secstorage/internal/server/storage/resource/model"
)

func extractUserId(ctx context.Context) model.UserId {
	return ctx.Value("userId").(model.UserId)
}
