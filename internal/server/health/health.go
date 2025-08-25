package management

import (
	healthpb "github.com/htranq/vortech-ome/api/v1/health"
)

func New() healthpb.HealthServer {
	return &healthServer{}
}

type healthServer struct {
	healthpb.UnimplementedHealthServer
}
