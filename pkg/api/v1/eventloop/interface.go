package eventloop

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

// common interface for event loops
type EventLoop interface {
	Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error)
}

type SimpleEventLoop interface {
	Run(ctx context.Context) (<-chan error, error)
}
