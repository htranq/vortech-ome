package streamtoken

import (
	"context"
	"errors"
	"time"

	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/internal/streamtoken/signer"
	"github.com/htranq/vortech-ome/pkg/config"
)

const (
	TokenQueryUrl = "stream_token"
	_tokenType    = "access_token"
)

type StreamToken interface {
	Issue(ctx context.Context, identity string) (string, error)
	Verify(ctx context.Context, raw string) error
}

type streamTokenImpl struct {
	cfg    *config.StreamToken
	signer signer.Signer
}

func New(cfg *config.StreamToken) (StreamToken, error) {
	if !cfg.GetEnabled() {
		logging.Logger(context.Background()).Fatal("StreamToken is disabled, init noop")
		return &noop{}, nil
	}

	sig, err := signer.New(_tokenType, cfg.GetJwtSigning())
	if err != nil {
		return nil, err
	}

	return &streamTokenImpl{
		cfg:    cfg,
		signer: sig,
	}, nil
}

func (s *streamTokenImpl) Issue(_ context.Context, identity string) (string, error) {
	token, err := s.signer.Create(identity, "")
	if err != nil {
		return "", err
	}

	return token.Raw, nil
}

func (s *streamTokenImpl) Verify(_ context.Context, raw string) error {
	token, err := s.signer.Parse(raw)
	if err != nil {
		return err
	}

	exp, err := token.GetExpirationTime()
	if err != nil {
		return err
	}

	if exp.Before(time.Now()) {
		return errors.New("token is expired")
	}

	return nil
}
