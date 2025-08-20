package client

import (
	"google.golang.org/grpc"

	pb "github.com/htranq/vortech-ome/pkg/cas/api/v1/table"
	configpb "github.com/htranq/vortech-ome/pkg/config"
	grpccli "github.com/htranq/vortech-ome/pkg/grpc"
)

type TableClient struct {
	pb.TableServiceClient
}

func NewTableClient(socket *configpb.TCPSocket, options ...grpc.DialOption) (*TableClient, error) {
	conn, err := grpccli.NewConnection(socket, options...)
	if err != nil {
		return nil, err
	}
	return &TableClient{
		TableServiceClient: pb.NewTableServiceClient(conn),
	}, nil
}
