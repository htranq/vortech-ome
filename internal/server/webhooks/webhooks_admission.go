package webhooks

import (
	"context"
	"errors"
	"net/url"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	webhookspb "github.com/htranq/vortech-ome/api/v1/webhooks"
	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/internal/streamtoken"
)

func (s *webhooksServer) Admission(ctx context.Context, request *webhookspb.AdmissionRequest) (*webhookspb.AdmissionResponse, error) {
	logger := logging.Logger(ctx).With(
		zap.Any("request_client", request.GetClient()),
		zap.Any("request_info", request.GetRequest()),
	)
	// Set response headers
	header := metadata.Pairs("Connection", "closed")
	_ = grpc.SendHeader(ctx, header)

	info := request.GetRequest()
	response := &webhookspb.AdmissionResponse{
		Allowed: true,
		Reason:  "authorized",
	}

Verifier:
	// Handle based on status
	switch info.GetStatus() {
	case _closing:
		logger.Info("admission handled with empty response")
		return &webhookspb.AdmissionResponse{}, nil

	case _opening:
		// only check for case outgoing : A client requests to play a stream
		if info.GetDirection() != _outgoing {
			response.Reason = "pass, only check for outgoing"
			break Verifier
		}

		// parse url
		u, err := url.Parse(info.GetUrl())
		if err != nil {
			response.Allowed = false
			response.Reason = err.Error()
			break Verifier
		}

		// verify token
		err = s.streamToken.Verify(ctx, u.Query().Get(streamtoken.TokenQueryUrl))
		if err != nil {
			response.Allowed = false
			response.Reason = err.Error()
		}

	default:
		logger.Error("unknown status")
		return nil, errors.New("unknown status")
	}

	logger.Info("admission handled successfully", zap.Any("response", response))

	return response, nil
}
