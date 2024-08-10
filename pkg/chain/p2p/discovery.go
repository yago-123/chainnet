package p2p

import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/sirupsen/logrus"
)

const (
	DiscoveryServiceTag = "node-p2p-discovery"
	DiscoveryTimeout    = 10 * time.Second
)

// discoveryNotifee handles peer discovery
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

// Discovery will be used to discover peers in the network and connect to them (these must be
// the methods used by chain package
type Discovery interface {
	Start() error
	Stop() error
}

type MdnsDiscovery struct {
	isActive bool
	mdns     mdns.Service
}

// NewMDNSDiscovery creates a new mDNS discovery service
func NewMdnsDiscovery(cfg *config.Config, host host.Host, netSubject observer.NetSubject) (*MdnsDiscovery, error) {
	mdnsService := mdns.NewMdnsService(host, DiscoveryServiceTag, newDiscoNotifee(cfg, host, netSubject))

	return &MdnsDiscovery{
		mdns:     mdnsService,
		isActive: false,
	}, nil
}

func (m *MdnsDiscovery) Start() error {
	if m.isActive {
		return nil
	}

	err := m.mdns.Start()
	if err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	m.isActive = true
	return nil
}

func (m *MdnsDiscovery) Stop() error {
	if !m.isActive {
		return nil
	}

	err := m.mdns.Close()
	if err != nil {
		return fmt.Errorf("failed to stop mDNS service: %w", err)
	}

	m.isActive = false

	return nil
}
