package discovery

import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/sirupsen/logrus"
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

// discoveryNotifee handles peer discovery logic at application level
type discoveryNotifee struct {
	host       host.Host
	netSubject observer.NetSubject
	logger     *logrus.Logger
}

func newDiscoNotifee(cfg *config.Config, host host.Host, netSubject observer.NetSubject) *discoveryNotifee {
	return &discoveryNotifee{
		host:       host,
		netSubject: netSubject,
		logger:     cfg.Logger,
	}
}

// HandlePeerFound connects to newly discovered peers
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.logger.Infof("discovered new peer %s", pi.ID)
	n.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

	ctx, cancel := context.WithTimeout(context.Background(), DiscoveryTimeout)
	defer cancel()

	if err := n.host.Connect(ctx, pi); err != nil {
		n.logger.Debugf("failed to connect to peer %s: %s", pi.ID, err)
		return
	}

	n.logger.Debugf("successfully connected to peer %s", pi.ID)

	n.netSubject.NotifyNodeDiscovered(string(pi.ID))
}
