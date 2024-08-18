package p2p

import (
	"chainnet/config"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/observer"
	"chainnet/pkg/p2p/discovery"
	"chainnet/pkg/p2p/pubsub"
	"chainnet/pkg/storage"
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
)

const (
	P2PObserverID = "p2p-observer"

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
	encoder  encoding.Encoding
	explorer *explorer.Explorer

	// bufferSize represents size of buffer for reading over the network
	bufferSize uint

	logger *logrus.Logger
}

func NewNodeP2P(
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
	pubsub, err := pubsub.NewGossipPubSub(ctx, host, encoder, netSubject, []string{}, true)
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
		bufferSize: cfg.P2PBufferSize,
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
	timeoutStream, err := NewTimeoutStream(ctx, n.cfg, n.host, peerID, AskLastHeaderProtocol)
	if err != nil {
		return nil, err
	}
	defer timeoutStream.Close()

	// read and decode reply
	data, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream %s: %w", timeoutStream.stream.ID(), err)
	}

	return n.encoder.DeserializeHeader(data)
}

// handleAskLastHeader handler that replies to the requests from AskLastHeader
func (n *NodeP2P) handleAskLastHeader(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, n.cfg)
	defer timeoutStream.Close()

	// get last block header
	header, err := n.explorer.GetLastHeader()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			n.logger.Infof("unable to retrieve last header for stream %s: no headers in the chain", stream.ID())
			return
		}

		n.logger.Errorf("error getting last block header for stream %s: %s", stream.ID(), err)
		return
	}

	// encode block header
	data, err := n.encoder.SerializeHeader(*header)
	if err != nil {
		n.logger.Errorf("error serializing block header for stream %s: %s", stream.ID(), err)
		return
	}

	// send block header to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		n.logger.Errorf("error writing block header for stream %s: %s", stream.ID(), err)
		return
	}
}

// AskSpecificBlock sends a request to a specific peer to get a block by hash
func (n *NodeP2P) AskSpecificBlock(ctx context.Context, peerID peer.ID, hash []byte) (*kernel.Block, error) {
	// open stream to peer with timeout
	timeoutStream, err := NewTimeoutStream(ctx, n.cfg, n.host, peerID, AskSpecificBlockProtocol)
	if err != nil {
		return nil, err
	}
	defer timeoutStream.Close()

	// write block hash required to stream
	_, err = timeoutStream.WriteWithTimeout(hash)
	if err != nil {
		return nil, fmt.Errorf("error writing block hash %x to stream: %w", hash, err)
	}
	// close write side of the stream so the peer knows we are done writing
	err = timeoutStream.stream.CloseWrite()
	if err != nil {
		return nil, fmt.Errorf("error closing write side of the stream: %w", err)
	}

	// read and decode block retrieved
	data, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeBlock(data)
}

// handleAskSpecificBlock handler that replies to the requests from AskSpecificBlock
func (n *NodeP2P) handleAskSpecificBlock(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, n.cfg)
	defer timeoutStream.Close()

	// read hash of block that is being requested
	hash, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		n.logger.Errorf("error reading block hash from stream %s: %s", stream.ID(), err)
		return
	}

	// retrieve block from explorer
	block, err := n.explorer.GetBlockByHash(hash)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			n.logger.Infof("unable to retrieve block for stream %s: block not found", stream.ID())
			return
		}

		n.logger.Errorf("error getting block with hash %x for stream %s: %s", hash, stream.ID(), err)
		return
	}

	// encode block
	data, err := n.encoder.SerializeBlock(*block)
	if err != nil {
		n.logger.Errorf("error serializing block with hash %x for stream %s: %s", hash, stream.ID(), err)
		return
	}

	// send block encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		n.logger.Errorf("error writing block with hash %x to stream %s: %s", hash, stream.ID(), err)
		return
	}
}

// AskAllHeaders sends a request to a specific peer to get all headers from the remote chain. The reply contains
// a list of headers unsorted
func (n *NodeP2P) AskAllHeaders(ctx context.Context, peerID peer.ID) ([]*kernel.BlockHeader, error) {
	// open stream to peer with timeout
	timeoutStream, err := NewTimeoutStream(ctx, n.cfg, n.host, peerID, AskAllHeaders)
	if err != nil {
		return nil, err
	}
	defer timeoutStream.Close()

	// read and decode block headers retrieved
	data, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeHeaders(data)
}

// handleAskAllHeaders handler that replies to the requests from AskAllHeaders
func (n *NodeP2P) handleAskAllHeaders(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, n.cfg)
	defer timeoutStream.Close()

	// retrieve headers from explorer
	headers, err := n.explorer.GetAllHeaders()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			n.logger.Infof("unable to retrieve headers for stream %s: no headers in the chain", stream.ID())
		}

		n.logger.Errorf("error getting headers for stream %s: %s", stream.ID(), err)
		return
	}

	// encode headers
	data, err := n.encoder.SerializeHeaders(headers)
	if err != nil {
		n.logger.Errorf("error serializing headers for stream %s: %s", stream.ID(), err)
		return
	}

	// send headers encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		n.logger.Errorf("error writing headers for stream %s: %s", stream.ID(), err)
		return
	}
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
