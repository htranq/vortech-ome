package outsider

import (
	"github.com/htranq/vortech-ome/internal/outsider/castable"
	"github.com/htranq/vortech-ome/pkg/config"
)

type Outsider struct {
	CasTable castable.CasTable
}

func New(cfg *config.Config) (*Outsider, error) {
	ctb, err := castable.New(cfg.GetCasTable())
	if err != nil {
		return nil, err
	}

	return &Outsider{
		CasTable: ctb,
	}, nil
}
