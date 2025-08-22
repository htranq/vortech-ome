package server

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/internal/xcontext"
)

// NOTE: Naming Convention
// - Middleware: HTTP layer (gRPC-Gateway) - uses runtime.HandlerFunc signature
// - Interceptor: gRPC layer - uses grpc.UnaryHandler signature

func executionInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()

	// Inject request ID if not present (should be propagated from HTTP headers via gRPC-Gateway)
	if xcontext.GetRequestID(ctx) == "" {
		ctx, _ = xcontext.InjectRequestID(ctx)
	}

	// Execute the handler
	resp, err = handler(ctx, req)

	// Log execution time
	duration := time.Since(start)
	logging.Logger(ctx).Info("GRPC handled",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", duration),
		zap.Int64("duration_ns", duration.Nanoseconds()))

	return resp, err
}

// executionMiddleware is a gRPC-Gateway middleware for logging request duration
func executionMiddleware(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		start := time.Now()

		// Create a context with request ID for HTTP requests
		ctx := r.Context()
		requestID := r.Header.Get(xcontext.XRequestIDHeader)
		if requestID == "" {
			ctx, requestID = xcontext.InjectRequestID(ctx)
			// Set the request ID as HTTP header so it propagates to gRPC
			r.Header.Set(xcontext.XRequestIDHeader, requestID)
			r = r.WithContext(ctx)
		}

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute the handler
		next(rw, r, pathParams)

		// Log HTTP request duration
		duration := time.Since(start)
		logging.Logger(ctx).Info("HTTP handled",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.Int64("duration_ns", duration.Nanoseconds()))
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
