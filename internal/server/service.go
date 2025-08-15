package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"buf.build/go/protovalidate"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/htranq/vortech-ome/internal/config"
	"github.com/htranq/vortech-ome/internal/logging"
	configpb "github.com/htranq/vortech-ome/pkg/config"
)

type Service interface {
	Logger() *zap.Logger
	GrpcListener() net.Listener
	GrpcServer() *grpc.Server
	Init(opts ...Option)
	Serve()
	Options() Options
}

type service struct {
	opts Options
}

func newService(cfg *configpb.Config, opts ...Option) Service {
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

	defaultOpts := []Option{
		WithLogger(logger),
		WithGrpcListener(listener),
		WithServerOptions(
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

	o := newOptions(append(defaultOpts, opts...)...)
	o.GrpcServer = grpc.NewServer(o.GrpcServerOptions...)

	svc := &service{
		opts: o,
	}

	return svc
}

func (s *service) Init(opts ...Option) {
	for _, opt := range opts {
		opt(&s.opts)
	}
}

func (s *service) Logger() *zap.Logger {
	return s.opts.Logger
}

func (s *service) GrpcListener() net.Listener {
	return s.opts.GrpcListener
}

func (s *service) GrpcServer() *grpc.Server {
	return s.opts.GrpcServer
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Serve() {
	logger := s.opts.Logger

	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			logger.Fatal("failed to exec before start", zap.Error(err))
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	s.serveGrpc(&wg)
	s.serveHttp(&wg)

	go s.watchShutdownSignal()

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			logger.Fatal("failed to exec before start", zap.Error(err))
		}
	}

	wg.Wait()
}

func (s *service) watchShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	sig := <-sigChan

	s.opts.Logger.Info(fmt.Sprint("got signal:", sig))
	s.opts.Logger.Info("start process before stop")
	failToStop := false
	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			failToStop = true
			s.opts.Logger.Error("failed to exec before stop")
		}
	}
	if failToStop {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func (s *service) serveGrpc(wg *sync.WaitGroup) {
	listener := s.opts.GrpcListener
	if listener == nil {
		return
	}
	wg.Add(1)
	logger := s.opts.Logger
	go func() {
		defer wg.Done()
		logger.Info("grpc listening", zap.String("address", listener.Addr().String()))
		if err := s.GrpcServer().Serve(listener); err != nil {
			logger.Fatal("failed to serveGrpc", zap.Error(err))
		}
	}()
}

func (s *service) serveHttp(wg *sync.WaitGroup) {
	listener := s.opts.HttpListener
	if listener == nil {
		return
	}
	wg.Add(1)
	//logger := s.opts.WithLogger
	go func() {
		defer wg.Done()
		// TODO serve http server
		//logger.Info("http listening", zap.String("address", listener.Addr().String()))
		//if err := http.Serve(listener, s.opts.HttpServerMux); err != nil {
		//	logger.Fatal("failed to serveHttp", zap.Error(err))
		//}
	}()
}

func LoadConfig(f *config.Flags) *configpb.Config {
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
