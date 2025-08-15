package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"buf.build/go/protovalidate"
	protovalidatemdw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	managementpb "github.com/htranq/vortech-ome/api/v1/management"
	"github.com/htranq/vortech-ome/internal/config"
	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/internal/server/api"
	managementsrv "github.com/htranq/vortech-ome/internal/server/management"
	configpb "github.com/htranq/vortech-ome/pkg/config"
)

func Run(f *config.Flags) {
	cfg := loadConfig(f)

	// TODO start/stop tracer

	serve(cfg)
}

func serve(cfg *configpb.Config) {
	var (
		ctx     = context.Background()
		service = newService(cfg, []api.Option{}...)
		logger  = service.Logger()
	)
	logger.Info("starting server", zap.Any("config", cfg))

	// init server handlers
	var (
		management = managementsrv.New()
	)

	// register grpc servers
	server := service.GrpcServer()
	managementpb.RegisterManagementServer(server, management)

	// register http servers
	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
				UseEnumNumbers:  false,
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.HasPrefix(key, "X-") || strings.HasPrefix(key, "x-") {
				return key, true
			}

			return key, false
		}),
	)

	service.HttpServerMux().Handle("/", grpcGatewayMux)
	if err := managementpb.RegisterManagementHandlerServer(ctx, grpcGatewayMux, management); err != nil {
		logger.Fatal("could not register management http server", zap.Error(err))
	}

	// register reflection
	reflection.Register(server)

	// Serve service
	service.Serve()
}

func loadConfig(f *config.Flags) *configpb.Config {
	// Use a temporary logger to parse the configuration and output.
	tmpLogger := logging.NewTmpLogger().With(zap.String("filename", f.ConfigPath))

	var cfg configpb.Config
	if err := config.ParseFile(f.ConfigPath, &cfg, f.Template); err != nil {
		tmpLogger.Fatal("parsing configuration failed", zap.Error(err))
	}

	if err := protovalidate.Validate(&cfg); err != nil {
		tmpLogger.Fatal("validating configuration failed", zap.Error(err))
	}

	if f.Validate {
		tmpLogger.Info("configuration validation was successful")
		os.Exit(0)
	}

	return &cfg
}

func newService(cfg *configpb.Config, opts ...api.Option) api.Service {
	err := logging.InitLogger(cfg.Logger)
	if err != nil {
		logging.NewTmpLogger().Error("fail to init logger", zap.Error(err))
	}

	// logger
	logger := logging.Logger(context.Background())

	// grpc listener
	grpcAddress := fmt.Sprintf("%s:%d",
		cfg.GetGrpcListener().GetTcp().GetAddress(),
		cfg.GetGrpcListener().GetTcp().GetPort())
	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		logger.Fatal("failed to new grpc listener", zap.Error(err))
	}

	// http listener
	httpAddress := fmt.Sprintf("%s:%d",
		cfg.GetHttpListener().GetTcp().GetAddress(),
		cfg.GetHttpListener().GetTcp().GetPort())
	httpListener, err := net.Listen("tcp", httpAddress)
	if err != nil {
		logger.Fatal("failed to new http listener", zap.Error(err))
	}

	// protovalidate validator
	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatal("fail to create validator", zap.Error(err))
	}

	defaultOpts := []api.Option{
		api.WithLogger(logger),
		api.WithGrpcListener(grpcListener),
		api.WithHttpListener(httpListener),
		api.WithServerOptions(
			grpc.UnaryInterceptor(protovalidatemdw.UnaryServerInterceptor(validator)),
		),
	}

	return api.NewService(append(defaultOpts, opts...)...)
}
