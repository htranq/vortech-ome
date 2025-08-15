package api

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	//healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

type Service interface {
	Logger() *zap.Logger
	Listener() net.Listener
	Server() *grpc.Server
	Init(opts ...Option)
	Serve()
	Options() Options
	HealthServer() *health.Server
}

type service struct {
	opts Options
}

func (s *service) Logger() *zap.Logger {
	return s.opts.Logger
}

func (s *service) Listener() net.Listener {
	return s.opts.Listener
}

func (s *service) Server() *grpc.Server {
	return s.opts.Server
}

func (s *service) HealthServer() *health.Server {
	return s.opts.HealthServer
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

	s.serve(&wg)
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

func (s *service) serve(wg *sync.WaitGroup) {
	listener := s.opts.Listener
	if listener == nil {
		return
	}
	wg.Add(1)
	logger := s.opts.Logger
	go func() {
		defer wg.Done()
		logger.Info("grpc listening", zap.String("address", listener.Addr().String()))
		if err := s.Server().Serve(listener); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()
}

func (s *service) serveHttp(wg *sync.WaitGroup) {
	listener := s.opts.HttpListener
	if listener == nil {
		return
	}
	wg.Add(1)
	//logger := s.opts.Logger
	go func() {
		defer wg.Done()
		// TODO serve http server
		//logger.Info("http listening", zap.String("address", listener.Addr().String()))
		//if err := http.Serve(listener, s.opts.HttpServerMux); err != nil {
		//	logger.Fatal("failed to serve", zap.Error(err))
		//}
	}()
}

func (s *service) Init(opts ...Option) {
	for _, opt := range opts {
		opt(&s.opts)
	}
}

func NewService(opts ...Option) Service {
	o := newOptions(opts...)

	o.Server = grpc.NewServer(o.ServerOptions...)
	if o.HealthServer != nil {
		// TODO: Register health server
		//healthv1.RegisterHealthServer(o.Server, o.HealthServer)
	}

	s := &service{
		opts: o,
	}

	return s
}
