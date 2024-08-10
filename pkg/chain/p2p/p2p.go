package p2p

import (
	"bufio"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/chain/p2p/discovery"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"

	"chainnet/config"
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

const (
	EchoProtocol           = "/echo/1.0.0"
	P2PComunicationTimeout = 10 * time.Second
)

type NodeP2P struct {
	cfg  *config.Config
	host host.Host

	// netSubject notifies other components about network events
	netSubject observer.NetSubject
	ctx        context.Context

	disco discovery.Discovery

	logger *logrus.Logger
}

func NewP2PNode(ctx context.Context, cfg *config.Config, netSubject observer.NetSubject) (*NodeP2P, error) {
	// create connection manager
	connMgr, err := connmgr.NewConnManager(
		int(cfg.P2PMinNumConn),
		int(cfg.P2PMaxNumConn),
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

	cfg.Logger.Debugf("p2p addresses:")
	for _, addr := range host.Addrs() {
		cfg.Logger.Debugf(" - %v\n", addr)
	}

	disco, err := discovery.NewMdnsDiscovery(cfg, host, netSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery module: %w", err)
	}

	return &NodeP2P{
		cfg:        cfg,
		host:       host,
		netSubject: netSubject,
		ctx:        ctx,
		disco:      disco,
		logger:     cfg.Logger,
	}, nil
}

func (n *NodeP2P) Start() error {
	return n.disco.Start()
}

func (n *NodeP2P) Stop() error {
	if err := n.disco.Stop(); err != nil {
		return err
	}

	return n.host.Close()
}

func (n *NodeP2P) InitHandlers() {
	n.host.SetStreamHandler(EchoProtocol, n.handleEchoStream)
}

func (n *NodeP2P) SendHello(peerID string) error {
	peerReference, err := peer.Decode(peerID)
	if err != nil {
		n.logger.Errorf("failed to decode peer reference: %v", err)
	}

	ctx, cancel := context.WithTimeout(n.ctx, P2PComunicationTimeout)
	defer cancel()

	stream, err := n.host.NewStream(ctx, peerReference, EchoProtocol)
	if err != nil {
		return fmt.Errorf("error enabling stream: %w", err)
	}
	defer stream.Close()

	// Send a message to the peer
	_, err = stream.Write([]byte("Hello from " + peerID + "!\n"))
	if err != nil {
		n.logger.Errorf("Failed to send message to peer %s: %s\n", peerID, err)
	}

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
