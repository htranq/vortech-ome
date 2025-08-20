package management

import (
	"context"

	healthpb "github.com/htranq/vortech-ome/api/v1/health"
)

func (*healthServer) GetStatus(_ context.Context, _ *healthpb.GetStatusRequest) (*healthpb.GetStatusResponse, error) {
	return &healthpb.GetStatusResponse{
		Status: "OK",
	}, nil
}
