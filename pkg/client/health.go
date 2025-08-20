package client

import (
	"google.golang.org/grpc"

	pb "github.com/htranq/vortech-ome/api/v1/health"
	configpb "github.com/htranq/vortech-ome/pkg/config"
	grpccli "github.com/htranq/vortech-ome/pkg/grpc"
)

type HealthClient struct {
	pb.HealthClient
}

func NewHealthClient(socket *configpb.TCPSocket, options ...grpc.DialOption) (*HealthClient, error) {
	conn, err := grpccli.NewConnection(socket, options...)
	if err != nil {
		return nil, err
	}
	return &HealthClient{
		HealthClient: pb.NewHealthClient(conn),
	}, nil
}
