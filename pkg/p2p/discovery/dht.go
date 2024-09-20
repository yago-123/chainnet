package discovery

import (
	"chainnet/config"
	"context"
	"fmt"

	ds "github.com/ipfs/go-datastore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	DHTDiscoveryType = "DHT"
)

type DHTDiscovery struct {
	dht      *dht.IpfsDHT
	isActive bool
}

func NewDHTDiscovery(cfg *config.Config, host host.Host) (*DHTDiscovery, error) {
	// todo(): consider adding persistent data store
	// todo(): roam around the options available for the DHT initialization
	// todo(): add seed nodes to the DHT via options too
	d := dht.NewDHT(context.Background(), host, ds.NewMapDatastore())

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

func (d *DHTDiscovery) Type() string {
	return DHTDiscoveryType
}
