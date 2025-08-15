package management

import (
	"context"

	managementpb "github.com/htranq/vortech-ome/api/v1/management"
)

func (*managementServer) GetStatus(_ context.Context, _ *managementpb.GetStatusRequest) (*managementpb.GetStatusReply, error) {
	return &managementpb.GetStatusReply{
		Status: "OK",
	}, nil
}
