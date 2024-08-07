package p2p

import (
	"chainnet/config"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
)

// Discovery will be used to discover peers in the network and connect to them (these must be
// the methods used by chain package
type Discovery interface {
}

const DiscoveryServiceTag = "node-p2p-discovery"

type NodeP2P struct {
}

// discoveryNotifee handles peer discovery
type discoveryNotifee struct {
	h      host.Host
	logger *logrus.Logger
}

func newDiscoNotifee(cfg *config.Config, host host.Host) *discoveryNotifee {
	return &discoveryNotifee{
		h:      host,
		logger: cfg.Logger,
	}
}

// HandlePeerFound connects to newly discovered peers
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.logger.Infof("discovered new peer %s\n", pi.ID)
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

	if err := n.h.Connect(context.Background(), pi); err != nil {
		n.logger.Infof("failed to connect to peer %s: %s\n", pi.ID, err)
	} else {
		n.logger.Infof("successfully connected to peer %s\n", pi.ID)
	}
}

func NewP2PNodeDiscovery(cfg *config.Config) (*NodeP2P, error) {
	connMgr, err := connmgr.NewConnManager(
		int(cfg.P2PMinNumConn),
		int(cfg.P2PMaxNumConn),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager during peer discovery: %w", err)
	}

	host, err := libp2p.New(
		libp2p.ConnectionManager(connMgr),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create host during peer discovery: %w", err)
	}

	cfg.Logger.Infof("Host created for peer discovery: %s", host.ID())

	cfg.Logger.Infof("Our addresses:")
	for _, addr := range host.Addrs() {
		cfg.Logger.Infof(" - %v\n", addr)
	}

	// set up mDNS discovery
	mdnsService := mdns.NewMdnsService(host, DiscoveryServiceTag, newDiscoNotifee(cfg, host))
	defer mdnsService.Close()

	err = mdnsService.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mDNS service: %w", err)
	}

	// select {}

	return &NodeP2P{}, nil
}
