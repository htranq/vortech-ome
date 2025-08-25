package castable

import (
	"context"

	"github.com/htranq/vortech-ome/internal/logging"
	caspb "github.com/htranq/vortech-ome/pkg/cas/api/v1/table"
	cascli "github.com/htranq/vortech-ome/pkg/cas/client"
	configpb "github.com/htranq/vortech-ome/pkg/config"
)

type CasTable interface {
	GetPlaybackUrl(ctx context.Context, tableID string) (string, error)
}

type casTableImpl struct {
	cfg    *configpb.CasTable
	client *cascli.TableClient
}

func New(cfg *configpb.CasTable) (CasTable, error) {
	if !cfg.GetEnabled() {
		logging.Logger(context.Background()).Info("CasTable disabled, init noop")
		return &noop{}, nil
	}
	client, err := cascli.NewTableClient(cfg.GetSocket())
	if err != nil {
		return nil, err
	}

	return &casTableImpl{client: client}, nil
}

func (c *casTableImpl) GetPlaybackUrl(ctx context.Context, tableID string) (string, error) {
	resp, err := c.client.GetPlayBackUrl(ctx, &caspb.GetPlayBackUrlRequest{TableId: tableID})
	if err != nil {
		return "", err
	}

	return resp.GetPlayBackUrl(), nil
}
