package discovery

import (
	"chainnet/config"
	"chainnet/pkg/observer"
	"context"
	"fmt"
	ds "github.com/ipfs/go-datastore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
	"time"
)

type DHTDiscovery struct {
	dht      *dht.IpfsDHT
	isActive bool
}

func NewDHTDiscovery(cfg *config.Config, host host.Host, netSubject observer.NetSubject) (*DHTDiscovery, error) {
	// todo(): consider adding persistent data store
	d := dht.NewDHT(context.Background(), host, ds.NewMapDatastore())

	// todo(): add function for watching peerstore from host in order to monitor while developing
	go printPeersDiscovered(host, cfg.Logger)

	return &DHTDiscovery{
		dht:      d,
		isActive: false,
	}, nil
}

func (d *DHTDiscovery) Start() error {
	if d.isActive {
		return nil
	}

	err := d.dht.Bootstrap(context.Background())
	if err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	d.isActive = true
	return nil
}

func (d *DHTDiscovery) Stop() error {
	if !d.isActive {
		return nil
	}

	err := d.dht.Close()
	if err != nil {
		return fmt.Errorf("failed to stop DHT: %w", err)
	}

	d.isActive = false
	return nil
}

func printPeersDiscovered(h host.Host, logger *logrus.Logger) {
	ticker := time.NewTicker(20 * time.Second)

	for {
		select {
		case <-ticker.C:
			logger.Errorf("Peers discovered: %d", len(h.Peerstore().Peers()))
			for _, p := range h.Peerstore().Peers() {
				logger.Errorf("- peer %s", p.String())
			}
		}
	}

}
