package p2p

import (
	"chainnet/config"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/chain/p2p/discovery"
	"chainnet/pkg/chain/p2p/pubsub"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
)

const (
	P2PObserverID = "p2p-observer"

	MaxNumberOfHeaderRequest = 1000

	P2PTotalTimeout = 20 * time.Second
	P2PWriteTimeout = 10 * time.Second
	P2PReadTimeout  = 10 * time.Second

	AskLastHeaderProtocol    = "/askLastHeader/0.1.0"
	AskSpecificBlockProtocol = "/askSpecificBlock/0.1.0"
	AskAllHeaders            = "/askAllHeaders/0.1.0"
)

type NodeP2P struct {
	cfg  *config.Config
	host host.Host

	// netSubject notifies other components about network events
	netSubject observer.NetSubject
	ctx        context.Context

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery
	// pubsub is in charge of setting up the logic for data propagation
	pubsub pubsub.PubSub

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

	// initialize discovery module
	disco, err := discovery.NewMdnsDiscovery(cfg, host, netSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery module: %w", err)
	}

	// initialize pubsub module
	pubsub, err := pubsub.NewGossipPubSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	return &NodeP2P{
		cfg:        cfg,
		host:       host,
		netSubject: netSubject,
		ctx:        ctx,
		disco:      disco,
		pubsub:     pubsub,
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
	n.host.SetStreamHandler(AskSpecificBlockProtocol, n.handleAskSpecificBlock)
	n.host.SetStreamHandler(AskAllHeaders, n.handleAskAllHeaders)
}

// AskLastHeader sends a request to a specific peer to get the last block header
func (n *NodeP2P) AskLastHeader(ctx context.Context, peerID peer.ID) (*kernel.BlockHeader, error) {
	// open stream to peer with timeout
	timeoutStream, err := NewTimeoutStream(ctx, n.host, peerID, P2PReadTimeout, P2PWriteTimeout, AskLastHeaderProtocol)
	if err != nil {
		return nil, fmt.Errorf("error enabling stream: %w", err)
	}
	defer timeoutStream.Close()

	var data []byte
	// read and decode reply
	_, err = timeoutStream.ReadWithTimeout(data)
	if err != nil {
		// todo(): handle the case in which the last header does not exist yet
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeHeader(data)
}

// handleAskLastHeader handler that replies to the requests from AskLastHeader
func (n *NodeP2P) handleAskLastHeader(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, P2PReadTimeout, P2PWriteTimeout)
	defer timeoutStream.Close()

	// get last block header
	header, err := n.explorer.GetLastHeader()
	if err != nil {
		n.logger.Errorf("error getting last block header: %s", err)
		return
	}

	// encode block header
	data, err := n.encoder.SerializeHeader(*header)
	if err != nil {
		n.logger.Errorf("error serializing block header: %s", err)
		return
	}

	// send block header to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		n.logger.Errorf("error writing block header to stream: %s", err)
		return
	}
}

// AskSpecificBlock sends a request to a specific peer to get a block by hash
func (n *NodeP2P) AskSpecificBlock(ctx context.Context, peerID peer.ID, hash []byte) (*kernel.Block, error) {
	// open stream to peer with timeout
	timeoutStream, err := NewTimeoutStream(ctx, n.host, peerID, P2PReadTimeout, P2PWriteTimeout, AskSpecificBlockProtocol)
	if err != nil {
		return nil, fmt.Errorf("error enabling stream: %w", err)
	}
	defer timeoutStream.Close()

	// write block hash required to stream
	_, err = timeoutStream.WriteWithTimeout(hash)
	if err != nil {
		return nil, fmt.Errorf("error writing block hash %x to stream: %w", hash, err)
	}

	var data []byte
	// read and decode block retrieved
	_, err = timeoutStream.ReadWithTimeout(data)
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeBlock(data)
}

// handleAskSpecificBlock handler that replies to the requests from AskSpecificBlock
func (n *NodeP2P) handleAskSpecificBlock(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, P2PReadTimeout, P2PWriteTimeout)
	defer timeoutStream.Close()

	var hash []byte
	// read hash of block that is being requested
	_, err := timeoutStream.ReadWithTimeout(hash)
	if err != nil {
		n.logger.Errorf("error reading block hash from stream: %s", err)
		return
	}

	// retrieve block from explorer
	block, err := n.explorer.GetBlockByHash(hash)
	if err != nil {
		n.logger.Errorf("error getting block with hash %x: %s", hash, err)
		return
	}

	// encode block
	data, err := n.encoder.SerializeBlock(*block)
	if err != nil {
		n.logger.Errorf("error serializing block with hash %x: %s", hash, err)
		return
	}

	// send block encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		n.logger.Errorf("error writing block with hash %x to stream: %s", hash, err)
		return
	}
}

// AskAllHeaders sends a request to a specific peer to get all headers from the remote chain. The reply contains
// a list of headers unsorted
func (n *NodeP2P) AskAllHeaders(_ context.Context, _ peer.ID) ([]*kernel.BlockHeader, error) {
	return nil, nil
}

// handleAskAllHeaders handler that replies to the requests from AskAllHeaders
func (n *NodeP2P) handleAskAllHeaders(_ network.Stream) {

}

func (n *NodeP2P) ID() string {
	return P2PObserverID
}

// OnBlockAddition is triggered as part of the chain controller, this function is
// executed when a new block is added into the chain
func (n *NodeP2P) OnBlockAddition(_ *kernel.Block) {
	// use n.pubsub to notify about new block addition
	// go n.NotifyBlockAddition(b)
}
