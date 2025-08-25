package castable

import (
	"context"

	"go.uber.org/zap"

	"github.com/htranq/vortech-ome/internal/logging"
)

type noop struct {
	playbackUrl string
}

func (s *noop) GetPlaybackUrl(ctx context.Context, tableID string) (string, error) {
	logging.Logger(ctx).Warn("noop GetPlaybackUrl, return default value",
		zap.String("table_id", tableID),
		zap.String("playback_url", s.playbackUrl))

	return s.playbackUrl, nil
}
