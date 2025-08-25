package management

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	managementpb "github.com/htranq/vortech-ome/api/v1/management"
	"github.com/htranq/vortech-ome/internal/logging"
)

const (
	_5minutes         = 5 * time.Minute
	_streamTokenParam = "stream_token"
)

func (s *managementServer) GetPlaybackUrl(ctx context.Context, request *managementpb.GetPlaybackUrlRequest) (*managementpb.GetPlaybackUrlResponse, error) {
	logger := logging.Logger(ctx).With(
		zap.String("table_id", request.GetTableId()),
		zap.String("service_id", request.GetServiceId()),
		zap.String("user_id", request.GetUserId()))

	// check authorization for this request
	reqTime := time.UnixMilli(request.GetAuthorization().GetTimestamp())
	if reqTime.Before(time.Now().Add(-_5minutes)) {
		logger.Error("authorization took too long to get playback url")
		return nil, status.Error(codes.FailedPrecondition, "Requested authorization time is in the past (more than 5 minutes)")
	}

	err := s.auth.Verify(toCanonicalString(request), request.GetAuthorization().GetSignature())
	if err != nil {
		logger.Error("failed to verify signature", zap.Error(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	//
	baseUrl, err := s.outsider.CasTable.GetPlaybackUrl(ctx, request.GetTableId())
	if err != nil {
		logger.Error("failed to get playback url from CAS Table", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	token, err := s.streamToken.Issue(request.GetUserId())
	if err != nil {
		logger.Error("failed to issue token", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		logger.Error("failed to parse playback url", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	query := u.Query()
	query.Set(_streamTokenParam, token)
	u.RawQuery = query.Encode()

	return &managementpb.GetPlaybackUrlResponse{
		Url:         u.String(),
		StreamToken: token,
	}, nil
}

func toCanonicalString(req *managementpb.GetPlaybackUrlRequest) string {
	return fmt.Sprintf("%s|%s|%s", req.GetTableId(), req.GetServiceId(), req.GetUserId())
}
