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
	_tokenType = "stream_token"
)

type StreamToken interface {
	Issue(identity string) (string, error)
	Verify(raw string) (bool, error)
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

func (s *streamTokenImpl) Issue(identity string) (string, error) {
	token, err := s.signer.Create(identity, "")
	if err != nil {
		return "", err
	}

	return token.Raw, nil
}

func (s *streamTokenImpl) Verify(raw string) (bool, error) {
	token, err := s.signer.Parse(raw)
	if err != nil {
		return false, err
	}

	exp, err := token.GetExpirationTime()
	if err != nil {
		return false, err
	}

	if exp.Before(time.Now()) {
		return false, errors.New("token is expired")
	}

	return true, nil
}
