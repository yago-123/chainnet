package discovery

import (
	"chainnet/config"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/sirupsen/logrus"
)

const (
	MDNSDiscoveryType = "mDNS"
)

type MdnsDiscovery struct {
	isActive bool
	mdns     mdns.Service
}

// NewMdnsDiscovery creates a new mDNS discovery service
func NewMdnsDiscovery(cfg *config.Config, host host.Host) (*MdnsDiscovery, error) {
	// inject the disco notifee logic into the MDNs algorithm
	mdnsService := mdns.NewMdnsService(host, DiscoveryServiceTag, newMDNSNotifee(host, cfg.Logger))

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

func (m *MdnsDiscovery) Type() string {
	return MDNSDiscoveryType
}

type notifee struct {
	host   host.Host
	logger *logrus.Logger
}

func newMDNSNotifee(host host.Host, logger *logrus.Logger) notifee {
	return notifee{
		host:   host,
		logger: logger,
	}
}

func (n notifee) HandlePeerFound(pi peer.AddrInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), DiscoveryTimeout)
	defer cancel()

	// try to connect to the peer and add the peer to the peerstore given that MDNs does not do that by default.
	// This way we can the host event bus will emit the peer found event. This addition to the peer store is done
	// by default in the case of other discovery types (like DHT)
	err := n.host.Connect(ctx, pi)
	if err != nil {
		n.logger.Errorf("failed to connect to peer %s after mDNS discovery: %s", pi.ID, err)
		return
	}
}
