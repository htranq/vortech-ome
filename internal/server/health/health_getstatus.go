package management

import (
	"context"
	"fmt"

	healthpb "github.com/htranq/vortech-ome/api/v1/health"
)

func (*healthServer) GetStatus(_ context.Context, request *healthpb.GetStatusRequest) (*healthpb.GetStatusResponse, error) {
	return &healthpb.GetStatusResponse{
		Status:  "OK",
		Message: fmt.Sprintf("Hello %s.", request.GetName()),
	}, nil
}
