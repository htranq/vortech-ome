package authorization

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/htranq/vortech-ome/internal/logging"
	"github.com/htranq/vortech-ome/pkg/config"
)

type Authorization interface {
	Sign(canonical string) string
	Verify(canonical, incomingSig string) error
}

type authorization struct {
	cfg *config.Authorization
}

func New(cfg *config.Authorization) (Authorization, error) {
	if !cfg.GetEnabled() {
		logging.Logger(context.Background()).Warn("Authorization disabled, init noop")
		return &noop{}, nil
	}
	if cfg.GetSecretKey() == "" {
		return nil, errors.New("invalid secret key")
	}

	return &authorization{
		cfg: cfg,
	}, nil
}

// Sign generates a HMAC-SHA256 signature from a canonical request string
func (s *authorization) Sign(canonical string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.GetSecretKey()))
	mac.Write([]byte(canonical))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// Verify checks if the provided signature matches the expected signature
// generated from the canonical request string and secret.
func (s *authorization) Verify(canonical, incomingSig string) error {
	expectedSig := s.Sign(canonical)

	expectedBytes, err := base64.StdEncoding.DecodeString(expectedSig)
	if err != nil {
		return err
	}
	incomingBytes, err := base64.StdEncoding.DecodeString(incomingSig)
	if err != nil {
		return err
	}
	if !hmac.Equal(expectedBytes, incomingBytes) {
		return errors.New("signature mismatch")
	}

	return nil
}
