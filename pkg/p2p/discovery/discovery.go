package discovery

import (
	"time"
)

const (
	DiscoveryServiceTag = "node-p2p-discovery"
	DiscoveryTimeout    = 10 * time.Second
)

// Discovery will be used to discover peers in the network level and connect to them
type Discovery interface {
	Start() error
	Stop() error
}
