package management

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	healthpb "github.com/htranq/vortech-ome/api/v1/health"
	"github.com/htranq/vortech-ome/internal/logging"
)

func (*healthServer) GetStatus(ctx context.Context, request *healthpb.GetStatusRequest) (*healthpb.GetStatusResponse, error) {
	logging.Logger(ctx).Info("GetStatus", zap.Any("request", request))

	return &healthpb.GetStatusResponse{
		Status:  "OK",
		Message: fmt.Sprintf("Hello %s.", request.GetName()),
	}, nil
}
