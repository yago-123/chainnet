package p2p

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

// Discovery will be used to discover peers in the network and connect to them (these must be
// the methods used by chain package
type Discovery interface {
}

const (
	DiscoveryServiceTag = "node-p2p-discovery"
	DiscoveryTimeout    = 10 * time.Second
)

// DiscoveryNotifee handles peer discovery
type DiscoveryNotifee struct {
	host       host.Host
	netSubject observer.NetSubject
	logger     *logrus.Logger
}

func NewDiscoNotifee(cfg *config.Config, host host.Host, netSubject observer.NetSubject) *DiscoveryNotifee {
	return &DiscoveryNotifee{
		host:       host,
		netSubject: netSubject,
		logger:     cfg.Logger,
	}
}

// HandlePeerFound connects to newly discovered peers
func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
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
