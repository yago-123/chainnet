package p2p

import (
	"chainnet/config"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/observer"
	"chainnet/pkg/p2p/discovery"
	"chainnet/pkg/p2p/pubsub"
	"context"
	"fmt"

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
	disco, err := discovery.NewMdnsDiscovery(cfg, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery module: %w", err)
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

func (n *WalletP2P) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	return n.pubsub.NotifyTransactionAdded(ctx, tx)
}
