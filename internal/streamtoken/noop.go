package streamtoken

import (
	"context"

	"github.com/htranq/vortech-ome/internal/logging"
)

type noop struct {
}

func (*noop) Issue(ctx context.Context, _ string) (string, error) {
	logging.Logger(ctx).Warn("noop unsupported")

	return "", nil
}

func (*noop) Verify(ctx context.Context, _ string) error {
	logging.Logger(ctx).Warn("noop unsupported")

	return nil
}
