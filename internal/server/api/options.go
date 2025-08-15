package api

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Options struct {
	Name   string
	Logger *zap.Logger

	GrpcServer        *grpc.Server
	GrpcServerOptions []grpc.ServerOption
	GrpcListener      net.Listener
	HttpListener      net.Listener

	BeforeStart []func() error
	AfterStart  []func() error
	BeforeStop  []func() error
}

type Option func(o *Options)

func Logger(logger *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func GrpcListener(listener net.Listener) Option {
	return func(o *Options) {
		o.GrpcListener = listener
	}
}

func ServerOptions(opts ...grpc.ServerOption) Option {
	return func(o *Options) {
		o.GrpcServerOptions = append(o.GrpcServerOptions, opts...)
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
