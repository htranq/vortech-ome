package server

import (
	"context"
	"fmt"
	"net"
	"os"

	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/htranq/vortech-ome/internal/config"
	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/internal/server/api"
	configpb "github.com/htranq/vortech-ome/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(f *config.Flags) {
	cfg := loadConfig(f)

	Serve(cfg)
}

func newService(cfg *configpb.Config, opts ...api.Option) api.Service {
	err := logging.InitLogger(cfg.Logger)
	if err != nil {
		logging.NewTmpLogger().Error("fail to init logger", zap.Error(err))
	}

	// grpc listener
	address := fmt.Sprintf("%s:%d",
		cfg.GetGrpcListener().GetTcp().GetAddress(),
		cfg.GetGrpcListener().GetTcp().GetPort())
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logging.NewTmpLogger().Fatal("failed to new listener", zap.Error(err))
	}

	logger := logging.Logger(context.Background())

	defaultOpts := []api.Option{
		api.Logger(logger),
		api.GrpcListener(listener),
		api.ServerOptions(
			grpc.ChainUnaryInterceptor(
				grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
			),
		),
	}

	//httpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.HttpListener.GetTcp().Address, cfg.HttpListener.GetTcp().Port))
	//if err != nil {
	//	logging.NewTmpLogger().Fatal("failed to new http listener", zap.Error(err))
	//}
	//httpListenerOption := api.HttpListener(httpListener)
	//defaultOpts = append(defaultOpts, httpListenerOption)

	svc := api.NewService(append(defaultOpts, opts...)...)

	return svc
}

func loadConfig(f *config.Flags) *configpb.Config {
	// Use a temporary logger to parse the configuration and output.
	tmpLogger := logging.NewTmpLogger().With(zap.String("filename", f.ConfigPath))

	var cfg configpb.Config
	if err := config.ParseFile(f.ConfigPath, &cfg, f.Template); err != nil {
		tmpLogger.Fatal("parsing configuration failed", zap.Error(err))
	}

	// TODO validate the configuration
	//if err := cfg.Validate(); err != nil {
	//	tmpLogger.Fatal("validating configuration failed", zap.Error(err))
	//}

	if f.Validate {
		tmpLogger.Info("configuration validation was successful")
		os.Exit(0)
	}

	return &cfg
}
