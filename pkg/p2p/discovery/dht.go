package discovery

import (
	"context"
	"github.com/ipfs/go-datastore"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/observer"
)

const (
	MinPeers = 1
	MaxPeers = 10
)

type DHTDiscovery struct {
	dht *kaddht.IpfsDHT
}

func NewDHTDiscovery(ctx context.Context, cfg *config.Config, host host.Host, netSubject observer.NetSubject) *DHTDiscovery {
	// todo(): move from NewDHT to dht.New to avoid panics and handle better
	dht := kaddht.NewDHT(ctx, host, datastore.NewMapDatastore())
	if dht == nil {

	}

	return &DHTDiscovery{dht: dht}
}

func (dht *DHTDiscovery) Start() error {
	return nil
}

func (dht *DHTDiscovery) Stop() error {
	return nil
}
