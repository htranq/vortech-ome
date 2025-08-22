package xcontext

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const (
	XRequestIDHeader = "x-request-id"

	_requestIDPrefix = "internal"
)

func GetRequestID(ctx context.Context) string {
	requestIds := metadata.ValueFromIncomingContext(ctx, XRequestIDHeader)
	if len(requestIds) < 1 {
		return ""
	}

	return requestIds[0]
}

func InjectRequestID(ctx context.Context) (context.Context, string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(make(map[string]string))
	}

	requestID := generateRequestID()
	md.Set(XRequestIDHeader, requestID)

	return metadata.NewIncomingContext(ctx, md), requestID
}

func generateRequestID() string {
	return fmt.Sprintf("%s|%s", _requestIDPrefix, uuid.New().String())
}
