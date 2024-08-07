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

const DiscoveryServiceTag = "node-p2p-discovery"

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
	fmt.Printf("Discovered new peer %s\n", pi.ID)
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

	if err := n.h.Connect(context.Background(), pi); err != nil {
		fmt.Printf("Failed to connect to peer %s: %s\n", pi.ID, err)
	} else {
		fmt.Printf("Successfully connected to peer %s\n", pi.ID)
	}
}

func NewP2PNodeDiscovery(cfg *config.Config) error {
	connMgr, err := connmgr.NewConnManager(
		1,
		100,
	)

	if err != nil {
		return fmt.Errorf("failed to create connection manager during peer discovery: %s", err)
	}

	host, err := libp2p.New(
		libp2p.ConnectionManager(connMgr),
	)

	if err != nil {
		return fmt.Errorf("failed to create host during peer discovery: %s", err)
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
		return fmt.Errorf("failed to start mDNS service: %v", err)
	}

	select {}
}
