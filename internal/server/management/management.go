package management

import (
	managementpb "github.com/htranq/vortech-ome/api/v1/management"
)

func New() managementpb.ManagementServer {
	return &managementServer{}
}

type managementServer struct {
	managementpb.UnimplementedManagementServer
}
