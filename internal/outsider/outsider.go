package outsider

import (
	"context"

	caspb "github.com/htranq/vortech-ome/pkg/cas/api/v1/table"
	cascli "github.com/htranq/vortech-ome/pkg/cas/client"
)

type Outsider struct {
	casTableCli *cascli.TableClient
}

func New(casTableCli *cascli.TableClient) *Outsider {
	return &Outsider{
		casTableCli: casTableCli,
	}
}

func (s *Outsider) GetCasTablePlaybackUrl(ctx context.Context, tableID string) (string, error) {
	resp, err := s.casTableCli.GetPlayBackUrl(ctx, &caspb.GetPlayBackUrlRequest{TableId: tableID})
	if err != nil {
		return "", err
	}

	return resp.GetPlayBackUrl(), nil
}
