package p2p

import (
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/chain/p2p/discovery"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
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
	P2PObserverID = "p2p-observer"

	P2PComunicationTimeout = 10 * time.Second
	P2PReadTimeout         = 10 * time.Second

	AskLastHeaderProtocol = "/lastHeader/0.1.0"
)

type NodeP2P struct {
	cfg  *config.Config
	host host.Host

	// netSubject notifies other components about network events
	netSubject observer.NetSubject
	ctx        context.Context

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery

	// encoder contains the communication data serialization between peers
	encoder encoding.Encoding

	explorer *explorer.Explorer

	logger *logrus.Logger
}

func NewP2PNode(
	ctx context.Context,
	cfg *config.Config,
	netSubject observer.NetSubject,
	encoder encoding.Encoding,
	explorer *explorer.Explorer,
) (*NodeP2P, error) {
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
		encoder:    encoder,
		explorer:   explorer,
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
	n.host.SetStreamHandler(AskLastHeaderProtocol, n.handleAskLastHeader)
}

// AskLastHeader sends a request to a specific peer to get the last block header
func (n *NodeP2P) AskLastHeader(peerID peer.ID) (*kernel.BlockHeader, error) {
	ctx, cancel := context.WithTimeout(n.ctx, P2PComunicationTimeout)
	defer cancel()

	// open stream to peer
	stream, err := n.host.NewStream(ctx, peerID, AskLastHeaderProtocol)
	if err != nil {
		return nil, fmt.Errorf("error enabling stream: %w", err)
	}
	defer stream.Close()

	// set deadline so we don't wait forever
	if err = stream.SetReadDeadline(time.Now().Add(P2PReadTimeout)); err != nil {
		return nil, fmt.Errorf("error setting read deadline: %w", err)
	}

	var data []byte
	// read and decode reply
	_, err = stream.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeHeader(data)
}

func (n *NodeP2P) handleAskLastHeader(stream network.Stream) {
	defer stream.Close()

	// get last block header
	n.explorer.GetLastBlockHeader()

	// encode block header

	// send block header over the network
}

func (n *NodeP2P) ID() string {
	return P2PObserverID
}

// OnBlockAddition is triggered as part of the chain controller, this function is
// executed when a new block is added into the chain
func (n *NodeP2P) OnBlockAddition(_ *kernel.Block) {

	// todo(): notify the network about the new node that has been added

}
