package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/htranq/vortech-ome/internal/xcontext"
	configpb "github.com/htranq/vortech-ome/pkg/config"
)

var _logger = NewTmpLogger()

func NewLogger(cfg *configpb.Logger) (*zap.Logger, error) {
	var (
		c    zap.Config
		opts []zap.Option
	)
	if cfg.GetPretty() {
		c = zap.NewDevelopmentConfig()
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	} else {
		c = zap.NewProductionConfig()
	}

	level := zap.NewAtomicLevel()

	levelName := "INFO"
	if cfg.Level != configpb.Logger_UNSPECIFIED {
		levelName = cfg.Level.String()
	}

	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		return nil, fmt.Errorf("could not parse log level %s", cfg.Level.String())
	}
	c.Level = level

	return c.Build(opts...)
}

func InitLogger(cfg *configpb.Logger) (err error) {
	_logger, err = NewLogger(cfg)

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

	return lg
}

func injectXRequestID(logger *zap.Logger, ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}
	requestID := xcontext.GetRequestID(ctx)
	if requestID == "" {
		return logger
	}

	return logger.With(zap.String(xcontext.XRequestIDHeader, requestID))
}
