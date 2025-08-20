package castable

import (
	"context"

	"go.uber.org/zap"

	"github.com/htranq/vortech-ome/internal/logging"
)

const (
	_noopPlaybackUrl = "wss://localhost:3333/app/stream"
)

type noop struct {
}

func (c *noop) GetPlaybackUrl(ctx context.Context, tableID string) (string, error) {
	logging.Logger(ctx).Warn("noop GetPlaybackUrl, return default value",
		zap.String("table_id", tableID),
		zap.String("playback_url", _noopPlaybackUrl))

	return "", nil
}
