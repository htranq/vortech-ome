package api

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

type Options struct {
	Name          string
	Logger        *zap.Logger
	Server        *grpc.Server
	ServerOptions []grpc.ServerOption
	Listener      net.Listener
	HttpListener  net.Listener
	HealthServer  *health.Server
	BeforeStart   []func() error
	AfterStart    []func() error
	BeforeStop    []func() error
}

type Option func(o *Options)

func Logger(logger *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func Listener(listener net.Listener) Option {
	return func(o *Options) {
		o.Listener = listener
	}
}

func ServerOptions(opts ...grpc.ServerOption) Option {
	return func(o *Options) {
		o.ServerOptions = append(o.ServerOptions, opts...)
	}
}

func HealthServer(healthServer *health.Server) Option {
	return func(o *Options) {
		o.HealthServer = healthServer
	}
}

func BeforeStart(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStart = append(o.BeforeStart, fn)
	}
}

func AfterStart(fn func() error) Option {
	return func(o *Options) {
		o.AfterStart = append(o.AfterStart, fn)
	}
}

func BeforeStop(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStop = append(o.BeforeStop, fn)
	}
}

func newOptions(opts ...Option) Options {
	o := Options{}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
