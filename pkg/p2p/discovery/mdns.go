package discovery

import (
	"chainnet/config"
	"chainnet/pkg/observer"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type MdnsDiscovery struct {
	isActive bool
	mdns     mdns.Service
}

// NewMDNSDiscovery creates a new mDNS discovery service
func NewMdnsDiscovery(cfg *config.Config, host host.Host, netSubject observer.NetSubject) (*MdnsDiscovery, error) {
	// inject the disco notifee logic into the Mdns algorithm
	// todo(): check if we really need a notifier if we already subscribe to the event bus
	mdnsService := mdns.NewMdnsService(host, DiscoveryServiceTag, emptyNotifee{})

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

type emptyNotifee struct{}

func (e emptyNotifee) HandlePeerFound(_ peer.AddrInfo) {
	// Do nothing
}
