package api

import (
	"net"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Options struct {
	Name   string
	Logger *zap.Logger

	GrpcListener      net.Listener
	GrpcServer        *grpc.Server
	GrpcServerOptions []grpc.ServerOption

	HttpListener  net.Listener
	HttpServerMux *http.ServeMux

	BeforeStart []func() error
	AfterStart  []func() error
	BeforeStop  []func() error
}

type Option func(o *Options)

func newOptions(opts ...Option) Options {
	o := Options{}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func WithLogger(logger *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithGrpcListener(listener net.Listener) Option {
	return func(o *Options) {
		o.GrpcListener = listener
	}
}

func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(o *Options) {
		o.GrpcServerOptions = append(o.GrpcServerOptions, opts...)
	}
}

func WithHttpListener(listener net.Listener) Option {
	return func(o *Options) {
		o.HttpListener = listener
		o.HttpServerMux = http.NewServeMux()
	}
}

func WithBeforeStart(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStart = append(o.BeforeStart, fn)
	}
}

func WithAfterStart(fn func() error) Option {
	return func(o *Options) {
		o.AfterStart = append(o.AfterStart, fn)
	}
}

func WithBeforeStop(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStop = append(o.BeforeStop, fn)
	}
}
