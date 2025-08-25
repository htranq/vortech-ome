package management

import (
	managementpb "github.com/htranq/vortech-ome/api/v1/management"
	"github.com/htranq/vortech-ome/internal/authorization"
	"github.com/htranq/vortech-ome/internal/outsider"
	"github.com/htranq/vortech-ome/internal/streamtoken"
)

func New(outsider *outsider.Outsider,
	authorization authorization.Authorization,
	streamToken streamtoken.StreamToken) managementpb.ManagementServer {
	return &managementServer{
		outsider:      outsider,
		authorization: authorization,
		streamToken:   streamToken,
	}
}

type managementServer struct {
	managementpb.UnimplementedManagementServer

	outsider *outsider.Outsider

	authorization authorization.Authorization
	streamToken   streamtoken.StreamToken
}
