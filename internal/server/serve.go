package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"buf.build/go/protovalidate"
	protovalidateitct "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	healthpb "github.com/htranq/vortech-ome/api/v1/health"
	managementpb "github.com/htranq/vortech-ome/api/v1/management"
	webhookspb "github.com/htranq/vortech-ome/api/v1/webhooks"
	authorizationsv "github.com/htranq/vortech-ome/internal/authorization"
	"github.com/htranq/vortech-ome/internal/config"
	"github.com/htranq/vortech-ome/internal/logging"
	outsidersv "github.com/htranq/vortech-ome/internal/outsider"
	"github.com/htranq/vortech-ome/internal/server/api"
	healthsrv "github.com/htranq/vortech-ome/internal/server/health"
	managementsrv "github.com/htranq/vortech-ome/internal/server/management"
	webhookssrv "github.com/htranq/vortech-ome/internal/server/webhooks"
	streamtokensv "github.com/htranq/vortech-ome/internal/streamtoken"
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
		service = newService(cfg)
		logger  = service.Logger()
	)
	logger.Info("starting server", zap.Any("config", cfg))
	logger.Info("init internal services")

	authorization, err := authorizationsv.New(cfg.GetAuthorization())
	if err != nil {
		logger.Fatal("failed to create authorization", zap.Error(err))
	}

	streamToken, err := streamtokensv.New(cfg.GetStreamToken())
	if err != nil {
		logger.Fatal("failed to create stream token", zap.Error(err))
	}

	logger.Info("init outsider")
	outsider, err := outsidersv.New(cfg)
	if err != nil {
		logger.Fatal("failed to create outsider", zap.Error(err))
	}

	// init server handlers
	var (
		health     = healthsrv.New()
		management = managementsrv.New(outsider, authorization, streamToken)
		webhooks   = webhookssrv.New(streamToken)
	)

	// register grpc servers
	grpcServer := service.GrpcServer()
	healthpb.RegisterHealthServer(grpcServer, health)
	managementpb.RegisterManagementServer(grpcServer, management)

	// create grpc-gateway mux
	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: false,
				UseEnumNumbers:  false,
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.HasPrefix(key, "x-") || strings.HasPrefix(key, "X-") {
				return key, true
			}
			return key, false
		}),
		runtime.WithMiddlewares(executionMiddleware),
	)

	// grpc-gateway has two approaches for HTTP-to-gRPC conversion:
	//
	// Option 1: RegisterHealthHandlerServer (Direct Server Handler)
	// HTTP -> JSON Parse -> Protobuf Convert -> Your Handler (DIRECT)
	// - ✅ Faster performance (~0.2ms latency)
	// - ✅ Lower memory usage
	// - ❌ Interceptors DON'T WORK (bypassed)
	// - ❌ Validation bypassed
	// - ❌ Different behavior than gRPC clients
	//
	// Option 2: RegisterHealthHandlerFromEndpoint (gRPC Client Call) ✅ CHOSEN
	// HTTP -> JSON Parse -> Protobuf Convert -> gRPC Call -> Interceptors -> Your Handler
	// - ❌ Slightly slower performance (~0.4ms latency, +0.2ms overhead)
	// - ❌ Higher memory usage (extra gRPC call)
	// - ✅ Interceptors WORK (injectRequestIDInterceptor, protovalidate)
	// - ✅ Validation works
	// - ✅ Consistent behavior with direct gRPC clients
	//
	// other:
	// Option 3: Native HTTP Handler (Direct Implementation) - MAXIMUM PERFORMANCE
	// HTTP -> Your HTTP Handler (DIRECT - NO CONVERSION)
	// - ✅ Fastest performance (~0.1ms latency, no conversion overhead)
	// - ✅ Lowest memory usage (no protobuf conversion)
	// - ✅ HTTP-specific features (streaming, custom headers, etc.)
	// - ❌ Code duplication (separate HTTP and gRPC implementations)
	// - ❌ Different validation logic needed
	// - ❌ Different middleware/interceptor patterns
	// - ❌ Inconsistent behavior between HTTP and gRPC clients
	//
	// Example implementation:
	// service.HttpServerMux().HandleFunc("/v1/health/status", func(w http.ResponseWriter, r *http.Request) {
	//     // Direct HTTP implementation - no gRPC conversion
	//     response := &healthpb.GetStatusResponse{Status: "healthy"}
	//     json.NewEncoder(w).Encode(response)
	// })

	grpcAddr := service.GrpcAddress()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// with HealthServer, we chose Option 2 because interceptor consistency is more valuable than micro-performance
	err = healthpb.RegisterHealthHandlerFromEndpoint(ctx, grpcGatewayMux, grpcAddr, dialOpts)
	if err != nil {
		logger.Fatal("could not register health http server", zap.Error(err))
	}

	// with WebhooksServer, we chose Option 1 to faster performance. Actually, Option 3 is the best
	err = webhookspb.RegisterWebhooksHandlerServer(ctx, grpcGatewayMux, webhooks)
	if err != nil {
		logger.Fatal("could not register webhooks http server", zap.Error(err))
	}

	// add grpc-gateway mux to HTTP server
	service.HttpServerMux().Handle("/", grpcGatewayMux)

	// register reflection
	reflection.Register(grpcServer)

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

func newService(cfg *configpb.Config) api.Service {
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

	opts := []api.Option{
		api.WithLogger(logger),
		api.WithGrpcListener(grpcListener),
		api.WithHttpListener(httpListener),
		api.WithServerOptions(
			grpc.ChainUnaryInterceptor(
				executionInterceptor,                                // FIRST - injects request ID and measures total time
				protovalidateitct.UnaryServerInterceptor(validator), // SECOND - validate request
			),
		),
	}

	return api.NewService(opts...)
}
