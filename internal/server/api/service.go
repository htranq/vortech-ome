package api

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service interface {
	Logger() *zap.Logger
	GrpcServer() *grpc.Server
	HttpServerMux() *http.ServeMux
	Serve()
}

type service struct {
	opts Options
}

func NewService(opts ...Option) Service {
	o := newOptions(opts...)
	o.GrpcServer = grpc.NewServer(o.GrpcServerOptions...)

	s := &service{
		opts: o,
	}

	return s
}

func (s *service) Logger() *zap.Logger {
	return s.opts.Logger
}

func (s *service) GrpcServer() *grpc.Server {
	return s.opts.GrpcServer
}

func (s *service) HttpServerMux() *http.ServeMux {
	return s.opts.HttpServerMux
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
	logger := s.opts.Logger
	go func() {
		defer wg.Done()
		logger.Info("http listening", zap.String("address", listener.Addr().String()))
		if err := http.Serve(listener, s.opts.HttpServerMux); err != nil {
			logger.Fatal("failed to serveHttp", zap.Error(err))
		}
	}()
}
