package castable

import (
	"context"

	caspb "github.com/htranq/vortech-ome/pkg/cas/api/v1/table"
	cascli "github.com/htranq/vortech-ome/pkg/cas/client"
)

type Castable interface {
	GetPlaybackUrl(ctx context.Context, tableID string) (string, error)
}

type castable struct {
	client *cascli.TableClient
}

func New(client *cascli.TableClient) Castable {
	if client == nil {
		return &noop{}
	}

	return &castable{client: client}
}

func (c *castable) GetPlaybackUrl(ctx context.Context, tableID string) (string, error) {
	resp, err := c.client.GetPlayBackUrl(ctx, &caspb.GetPlayBackUrlRequest{TableId: tableID})
	if err != nil {
		return "", err
	}

	return resp.GetPlayBackUrl(), nil
}
