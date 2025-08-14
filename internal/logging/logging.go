package logging

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	//"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/htranq/vortech-ome/pkg/config"
)

var (
	_logger           = NewTmpLogger()
	_xRequestIDHeader = "x-request-id"
)

func NewLogger(msg *config.Logger) (*zap.Logger, error) {
	var c zap.Config
	var opts []zap.Option
	if msg.GetPretty() {
		c = zap.NewDevelopmentConfig()
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	} else {
		c = zap.NewProductionConfig()
	}

	level := zap.NewAtomicLevel()

	levelName := "INFO"
	if msg.Level != config.Logger_UNSPECIFIED {
		levelName = msg.Level.String()
	}

	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		return nil, fmt.Errorf("could not parse log level %s", msg.Level.String())
	}
	c.Level = level

	return c.Build(opts...)
}

func InitLogger(msg *config.Logger) (err error) {
	_logger, err = NewLogger(msg)
	return err
}

func NewTmpLogger() *zap.Logger {
	c := zap.NewProductionConfig()
	c.DisableStacktrace = true
	l, err := c.Build()
	if err != nil {
		panic(err)
	}
	return l
}

// Logger Return new config with context value
// ctx:  nillable
func Logger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return _logger
	}
	lg := injectXRequestID(_logger, ctx)
	//lg = injectDatadogTracing(lg, ctx)

	return lg
}

func SetXRequestIDHeader(headerName string) {
	_xRequestIDHeader = headerName
}

//func injectDatadogTracing(config *zap.Logger, ctx context.Context) *zap.Logger {
//
//	if service, ok := os.LookupEnv("DD_SERVICE"); ok {
//		config = config.With(zap.String("dd.service", service))
//	}
//
//	if env, ok := os.LookupEnv("DD_ENV"); ok {
//		config = config.With(zap.String("dd.env", env))
//	}
//
//	if version, ok := os.LookupEnv("DD_VERSION"); ok {
//		config = config.With(zap.String("dd.version", version))
//	}
//
//	if ctx == nil {
//		return config
//	}
//	span, ok := tracer.SpanFromContext(ctx)
//	if !ok {
//		return config
//	}
//
//	spanCtx := span.Context()
//
//	return config.With(zap.String("dd.trace_id", strconv.FormatUint(spanCtx.TraceID(), 10)),
//		zap.String("dd.span_id", strconv.FormatUint(spanCtx.SpanID(), 10)))
//}

func injectXRequestID(logger *zap.Logger, ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}
	requestID := getRequestID(ctx)
	if requestID == "" {
		return logger
	}
	return logger.With(zap.String("x_request_id", requestID))
}

func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	requestIds := md.Get(_xRequestIDHeader)
	if len(requestIds) < 1 {
		return ""
	}
	return requestIds[0]
}
