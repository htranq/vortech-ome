package streamtoken

import (
	"github.com/htranq/vortech-ome/internal/streamtoken/signer"
	"github.com/htranq/vortech-ome/pkg/config"
)

type StreamToken interface {
}

type streamTokenImpl struct {
	cfg    *config.StreamToken
	signer signer.Signer
}

func New() StreamToken {
	return &streamTokenImpl{}
}

func (s *streamTokenImpl) Issue() (string, error) {
	// TODO
	return "", nil
}

func (s *streamTokenImpl) Verify(token string) (bool, error) {
	// TODO
	return true, nil
}
