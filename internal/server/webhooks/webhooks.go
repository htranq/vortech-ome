package webhooks

import (
	webhookspb "github.com/htranq/vortech-ome/api/v1/webhooks"
	"github.com/htranq/vortech-ome/internal/streamtoken"
)

const (
	// admission directions
	_incoming string = "incoming"
	_outgoing string = "outgoing"

	// admission statuses
	_opening string = "opening"
	_closing string = "closing"
)

func New(streamToken streamtoken.StreamToken) webhookspb.WebhooksServer {
	return &webhooksServer{
		streamToken: streamToken,
	}
}

type webhooksServer struct {
	webhookspb.UnimplementedWebhooksServer

	streamToken streamtoken.StreamToken
}
