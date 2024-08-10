package p2p

import (
	"bufio"
	"chainnet/pkg/chain/observer"
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

	netSubject observer.NetSubject

	ctx    context.Context
	logger *logrus.Logger
}

func NewP2PNode(ctx context.Context, cfg *config.Config, netSubject observer.NetSubject) (*NodeP2P, error) {
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
		cfg:        cfg,
		host:       host,
		netSubject: netSubject,
		ctx:        ctx,
		logger:     cfg.Logger,
	}, nil
}

func (n *NodeP2P) Start() error {
	// init node discovery
	return nil
}

func (n *NodeP2P) Stop() error {
	n.host.Addrs()
	return n.host.Close()
}

func (n *NodeP2P) InitHandlers() {
	n.host.SetStreamHandler("/echo/1.0.0", n.handleEchoStream)
}

// InitNodeDiscovery initializes the mechanism for discovering peers in the network
func (n *NodeP2P) InitNodeDiscovery() error {
	mdnsService := mdns.NewMdnsService(n.host, DiscoveryServiceTag, NewDiscoNotifee(n.cfg, n.host, n.netSubject))

	err := mdnsService.Start()
	if err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	// mdnsService can't be returned because is unexported from the package itself,
	// so we need to wait for closure in a goroutine
	go func() {
		<-n.ctx.Done()
		mdnsService.Close()
	}()

	return nil
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
