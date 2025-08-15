package client

import (
	"google.golang.org/grpc"

	pb "github.com/htranq/vortech-ome/api/v1/management"
	"github.com/htranq/vortech-ome/pkg/config"
	grpccli "github.com/htranq/vortech-ome/pkg/grpc"
)

type ManagementClient struct {
	pb.ManagementClient
}

func NewMyGSCClient(config *config.TCPSocket, options ...grpc.DialOption) (*ManagementClient, error) {
	conn, err := grpccli.NewConnection(config, options...)
	if err != nil {
		return nil, err
	}
	return &ManagementClient{
		ManagementClient: pb.NewManagementClient(conn),
	}, nil
}
