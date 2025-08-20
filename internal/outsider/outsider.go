package outsider

import (
	"github.com/htranq/vortech-ome/internal/outsider/castable"
	cascli "github.com/htranq/vortech-ome/pkg/cas/client"
	"github.com/htranq/vortech-ome/pkg/config"
)

type Outsider struct {
	castable castable.Castable
}

func New(cfg *config.Config) (*Outsider, error) {
	var (
		err    error
		casCli *cascli.TableClient
	)

	if cfg.GetCasTable().GetEnabled() {
		casCli, err = cascli.NewTableClient(cfg.GetCasTable().GetSocket())
		if err != nil {
			return nil, err
		}
	}

	return &Outsider{
		castable: castable.New(casCli),
	}, nil
}
