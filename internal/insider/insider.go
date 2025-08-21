package insider

import (
	"github.com/htranq/vortech-ome/internal/insider/authorization"
	"github.com/htranq/vortech-ome/pkg/config"
)

type Insider struct {
	authorization authorization.Authorization
}

func New(cfg *config.Config) (*Insider, error) {
	auth, err := authorization.New(cfg.GetAuthorization())
	if err != nil {
		return nil, err
	}

	return &Insider{
		authorization: auth,
	}, nil
}
