package eventloop

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

// common interface for event loops
type EventLoop interface {
	Run(namespaces *clients.NamespacesByResourceWatcher, opts clients.WatchOpts) (<-chan error, error)
}
