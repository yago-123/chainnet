package discovery

import (
	"chainnet/pkg/observer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
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

func NotifyPeersDiscovered(host host.Host, subject observer.NetSubject) error {
	sub, err := host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	if err != nil {
		return err
	}

	go func() {
		for evt := range sub.Out() {
			// Cast the event and extract peer ID
			peerEvt := evt.(event.EvtPeerIdentificationCompleted)
			subject.NotifyNodeDiscovered(peerEvt.Peer)
		}
	}()

	return nil
}
