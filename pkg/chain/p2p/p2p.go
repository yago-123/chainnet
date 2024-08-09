package p2p

import (
	"bufio"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"

	"chainnet/config"
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

type NodeP2P struct {
	cfg  *config.Config
	host host.Host

	ctx    context.Context
	logger *logrus.Logger
}

func NewP2PNodeDiscovery(ctx context.Context, cfg *config.Config) (*NodeP2P, error) {
	connMgr, err := connmgr.NewConnManager(
		int(cfg.P2PMinNumConn),
		int(cfg.P2PMaxNumConn),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager during peer discovery: %w", err)
	}

	host, err := libp2p.New(
		libp2p.ConnectionManager(connMgr),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create host during peer discovery: %w", err)
	}

	cfg.Logger.Debugf("host created for peer discovery: %s", host.ID())

	cfg.Logger.Debugf("p2p addresses:")
	for _, addr := range host.Addrs() {
		cfg.Logger.Debugf(" - %v\n", addr)
	}

	return &NodeP2P{
		cfg:    cfg,
		host:   host,
		ctx:    ctx,
		logger: cfg.Logger,
	}, nil
}

func (n *NodeP2P) Stop() error {
	n.host.Addrs()
	return n.host.Close()
}

func (n *NodeP2P) InitializeHandlers() {
	n.host.SetStreamHandler("/echo/1.0.0", n.handleEchoStream)
}

func (n *NodeP2P) Sync() error {
	// set up mDNS discovery
	mdnsService := mdns.NewMdnsService(n.host, DiscoveryServiceTag, NewDiscoNotifee(n.cfg, n.host))
	defer mdnsService.Close()

	err := mdnsService.Start()
	if err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	select { //nolint:gosimple // ignore linter in this case
	case <-n.ctx.Done():
		return n.ctx.Err()
	}
}

func (n *NodeP2P) handleEchoStream(stream network.Stream) {
	defer stream.Close()

	buf := bufio.NewReader(stream)
	for {
		str, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		n.logger.Debugf("Received: %s", str)
	}
}
