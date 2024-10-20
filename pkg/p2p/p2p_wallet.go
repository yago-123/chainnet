package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/p2p/discovery"
	"github.com/yago-123/chainnet/pkg/p2p/pubsub"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
)

type WalletP2P struct {
	cfg  *config.Config
	host host.Host

	ctx context.Context

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery
	// pubsub is in charge of setting up the logic for data propagation
	pubsub pubsub.PubSub
	// encoder contains the communication data serialization between peers
	encoder encoding.Encoding

	logger *logrus.Logger
}

func NewWalletP2P(
	ctx context.Context,
	cfg *config.Config,
	encoder encoding.Encoding,
) (*WalletP2P, error) {
	// create connection manager
	connMgr, err := connmgr.NewConnManager(
		int(cfg.P2P.MinNumConn), //nolint:gosec // this overflowing is OK
		int(cfg.P2P.MaxNumConn), //nolint:gosec // this overflowing is OK
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager during peer discovery: %w", err)
	}

	// create host
	host, err := libp2p.New(
		libp2p.ConnectionManager(connMgr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host during peer discovery: %w", err)
	}

	cfg.Logger.Debugf("host created for peer discovery: %s", host.ID())

	// initialize discovery module
	disco, err := discovery.NewDHTDiscovery(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT discovery module: %w", err)
	}

	// initialize pubsub module
	pubsub, err := pubsub.NewGossipPubSub(ctx, cfg, host, encoder, observer.NewNetSubject(), []string{}, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	return &WalletP2P{
		cfg:     cfg,
		host:    host,
		ctx:     ctx,
		disco:   disco,
		pubsub:  pubsub,
		encoder: encoder,
		logger:  cfg.Logger,
	}, nil
}

func (n *WalletP2P) Start() error {
	return n.disco.Start()
}

func (n *WalletP2P) Stop() error {
	if err := n.disco.Stop(); err != nil {
		return err
	}

	return n.host.Close()
}

// todo() remove duplication of this method between p2p_wallet and p2p_node
func (n *WalletP2P) ConnectToSeeds() error {
	for _, seed := range n.cfg.SeedNodes {
		addr, err := peer.AddrInfoFromString(
			fmt.Sprintf("/dns4/%s/tcp/%d/p2p/%s", seed.Address, seed.Port, seed.PeerID),
		)
		if err != nil {
			return fmt.Errorf("failed to parse multiaddress: %w", err)
		}

		err = n.host.Connect(n.ctx, *addr)
		if err != nil {
			n.cfg.Logger.Errorf("failed to connect to seed node %s: %v", addr, err)
			continue
		}

		n.cfg.Logger.Infof("connected to seed node %s", addr.ID.String())
	}

	return nil
}

func (n *WalletP2P) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	return n.pubsub.NotifyTransactionAdded(ctx, tx)
}
