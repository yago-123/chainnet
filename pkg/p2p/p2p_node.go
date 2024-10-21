package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"net/http"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus/util"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/p2p/discovery"
	"github.com/yago-123/chainnet/pkg/p2p/events"
	"github.com/yago-123/chainnet/pkg/p2p/pubsub"
	"github.com/yago-123/chainnet/pkg/storage"

	"github.com/libp2p/go-libp2p"
	p2pConfig "github.com/libp2p/go-libp2p/config"
	p2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
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

type nodeP2PHandler struct {
	logger   *logrus.Logger
	encoder  encoding.Encoding
	explorer *explorer.Explorer

	cfg *config.Config
}

func newNodeP2PHandler(cfg *config.Config, encoder encoding.Encoding, explorer *explorer.Explorer) *nodeP2PHandler {
	return &nodeP2PHandler{
		logger:   cfg.Logger,
		encoder:  encoder,
		explorer: explorer,
		cfg:      cfg,
	}
}

// handleAskLastHeader handler that replies to the requests from AskLastHeader
func (h *nodeP2PHandler) handleAskLastHeader(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, h.cfg)
	defer timeoutStream.Close()

	// get last block header
	header, err := h.explorer.GetLastHeader()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.logger.Infof("unable to retrieve last header for stream %s: no headers in the chain", stream.ID())
			return
		}

		h.logger.Errorf("error getting last block header for stream %s: %s", stream.ID(), err)
		return
	}

	// encode block header
	data, err := h.encoder.SerializeHeader(*header)
	if err != nil {
		h.logger.Errorf("error serializing block header for stream %s: %s", stream.ID(), err)
		return
	}

	// send block header to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		h.logger.Errorf("error writing block header for stream %s: %s", stream.ID(), err)
		return
	}
}

// handleAskSpecificBlock handler that replies to the requests from AskSpecificBlock
func (h *nodeP2PHandler) handleAskSpecificBlock(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, h.cfg)
	defer timeoutStream.Close()

	// read hash of block that is being requested
	hash, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		h.logger.Errorf("error reading block hash from stream %s: %s", stream.ID(), err)
		return
	}

	// retrieve block from explorer
	block, err := h.explorer.GetBlockByHash(hash)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.logger.Infof("unable to retrieve block for stream %s: block %x not found", stream.ID(), hash)
			return
		}

		h.logger.Errorf("error getting block with hash %x for stream %s: %s", hash, stream.ID(), err)
		return
	}

	// encode block
	data, err := h.encoder.SerializeBlock(*block)
	if err != nil {
		h.logger.Errorf("error serializing block with hash %x for stream %s: %s", hash, stream.ID(), err)
		return
	}

	// send block encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		h.logger.Errorf("error writing block with hash %x to stream %s: %s", hash, stream.ID(), err)
		return
	}
}

// handleAskAllHeaders handler that replies to the requests from AskAllHeaders
func (h *nodeP2PHandler) handleAskAllHeaders(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, h.cfg)
	defer timeoutStream.Close()

	// retrieve headers from explorer
	headers, err := h.explorer.GetAllHeaders()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.logger.Infof("unable to retrieve headers for stream %s: no headers in the chain", stream.ID())
		}

		h.logger.Errorf("error getting headers for stream %s: %s", stream.ID(), err)
		return
	}

	// encode headers
	data, err := h.encoder.SerializeHeaders(headers)
	if err != nil {
		h.logger.Errorf("error serializing headers for stream %s: %s", stream.ID(), err)
		return
	}

	// send headers encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		h.logger.Errorf("error writing headers for stream %s: %s", stream.ID(), err)
		return
	}
}

type HTTPRouter struct {
	r        *httprouter.Router
	encoder  encoding.Encoding
	explorer *explorer.Explorer

	cfg *config.Config
}

func NewHTTPRouter(cfg *config.Config, encoder encoding.Encoding, explorer *explorer.Explorer) *HTTPRouter {
	router := &HTTPRouter{
		r:        httprouter.New(),
		encoder:  encoder,
		explorer: explorer,
		cfg:      cfg,
	}

	// todo() move these paths to constants?
	router.r.GET("/address/:address/transactions", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router.listTransactions(w, r, ps)
	})
	router.r.GET("/address/:address/utxos", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router.listUTXOs(w, r, ps)
	})
	router.r.GET("/address/:address/balance", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router.getAddressBalance(w, r, ps)
	})

	return router
}

// todo(): add cancelling mechanism, improve error handling and use flags to pass port
func (router *HTTPRouter) Start() error {
	go http.ListenAndServe(":8080", router.r)

	return nil
}

// todo(): add functionality
func (router *HTTPRouter) Stop() error {
	return nil
}

func (router *HTTPRouter) listTransactions(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	// todo() replace this method with all transactions instead of only non-spent ones
	transactions, err := router.explorer.FindUnspentTransactions(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
	}

	// todo() replace encoder to use grpc
	err = json.NewEncoder(w).Encode(transactions) //nolint:musttag // not sure which encoding will use in the future
	if err != nil {
		http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
	}
}

func (router *HTTPRouter) listUTXOs(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	utxos, err := router.explorer.FindUnspentTransactions(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve utxos", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(utxos) //nolint:musttag // not sure which encoding will use in the future
	if err != nil {
		http.Error(w, "Failed to encode UTXOs", http.StatusInternalServerError)
	}
}

func (router *HTTPRouter) getAddressBalance(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	balanceResponse, err := router.explorer.CalculateAddressBalance(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to find unspent transactions", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(balanceResponse)
	if err != nil {
		http.Error(w, "Failed to encode balance", http.StatusInternalServerError)
	}
}

type NodeP2P struct {
	cfg  *config.Config
	host host.Host

	// netSubject notifies other components about network events
	netSubject observer.NetSubject
	ctx        context.Context

	// discoDHT is in charge of setting up the logic for remote node discovery
	discoDHT discovery.Discovery
	// discoMDNS is in charge of setting up the logic for local node discovery
	discoMDNS discovery.Discovery
	// pubsub is in charge of setting up the logic for data propagation
	pubsub pubsub.PubSub
	// encoder contains the communication data serialization between peers
	encoder  encoding.Encoding
	explorer *explorer.Explorer

	router *HTTPRouter

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
	// options represent the configuration options for the libp2p host
	options := []p2pConfig.Option{}

	// create connection manager
	connMgr, err := connmgr.NewConnManager(
		int(cfg.P2P.MinNumConn), //nolint:gosec // this overflowing is OK
		int(cfg.P2P.MaxNumConn), //nolint:gosec // this overflowing is OK
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager during peer discovery: %w", err)
	}

	// add connection manager and listening address to options
	options = append(options, libp2p.ConnectionManager(connMgr))
	options = append(options, libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.P2P.PeerPort)))

	// add identity if the keys exists
	if cfg.P2P.Identity.PrivKeyPath != "" {
		privKeyBytes, errKey := util.ReadECDSAPemPrivateKey(cfg.P2P.Identity.PrivKeyPath)
		if errKey != nil {
			return nil, fmt.Errorf("error reading private key: %w", errKey)
		}

		priv, errKey := util.ConvertBytesToECDSAPriv(privKeyBytes)
		if errKey != nil {
			return nil, fmt.Errorf("error converting private key: %w", errKey)
		}

		p2pKey, _, errKey := p2pCrypto.ECDSAKeyPairFromKey(priv)
		if errKey != nil {
			return nil, fmt.Errorf("error creating p2p key pair: %w", errKey)
		}

		// add peer identity to options
		options = append(options, libp2p.Identity(p2pKey))
	}

	// create host
	host, err := libp2p.New(
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host during peer discovery: %w", err)
	}

	cfg.Logger.Debugf("host created for peer discovery: %s", host.ID())

	// create listener for host events
	err = events.InitializeHostEventsSubscription(ctx, cfg.Logger, host, netSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize subscription for host events: %w", err)
	}

	// initialize DHT discovery module remote discovery
	discoDHT, err := discovery.NewDHTDiscovery(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT discovery module: %w", err)
	}

	// initialize MDNS discovery module for local discovery
	discoMDNS, err := discovery.NewMdnsDiscovery(cfg, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS discovery module: %w", err)
	}

	// initialize pubsub module
	topics := []string{pubsub.BlockAddedPubSubTopic, pubsub.TxAddedPubSubTopic}
	pubsub, err := pubsub.NewGossipPubSub(ctx, cfg, host, encoder, netSubject, topics, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	// initialize HTTP router for handling HTTP requests (wallet, information requests...)
	router := NewHTTPRouter(cfg, encoder, explorer)

	// initialize handlers
	handler := newNodeP2PHandler(cfg, encoder, explorer)
	host.SetStreamHandler(AskLastHeaderProtocol, handler.handleAskLastHeader)
	host.SetStreamHandler(AskSpecificBlockProtocol, handler.handleAskSpecificBlock)
	host.SetStreamHandler(AskAllHeaders, handler.handleAskAllHeaders)

	return &NodeP2P{
		cfg:        cfg,
		host:       host,
		netSubject: netSubject,
		ctx:        ctx,
		discoDHT:   discoDHT,
		discoMDNS:  discoMDNS,
		pubsub:     pubsub,
		encoder:    encoder,
		router:     router,
		explorer:   explorer,
		bufferSize: cfg.P2P.BufferSize,
		logger:     cfg.Logger,
	}, nil
}

func (n *NodeP2P) Start() error {
	if err := n.discoDHT.Start(); err != nil {
		return fmt.Errorf("failed to start DHT discovery: %w", err)
	}

	if err := n.discoMDNS.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS discovery: %w", err)
	}

	if err := n.router.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP router: %w", err)
	}

	return nil
}

func (n *NodeP2P) Stop() error {
	if err := n.discoDHT.Stop(); err != nil {
		return fmt.Errorf("error stopping DHT discovery: %w", err)
	}

	if err := n.discoMDNS.Stop(); err != nil {
		return fmt.Errorf("error stopping mDNS discovery: %w", err)
	}

	if err := n.router.Stop(); err != nil {
		return fmt.Errorf("error stopping HTTP router: %w", err)
	}

	return n.host.Close()
}

func (n *NodeP2P) ConnectToSeeds() error {
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

func (n *NodeP2P) ID() string {
	return P2PObserverID
}

// OnBlockAddition is triggered as part of the chain controller, this function is
// executed when a new block is added into the chain
func (n *NodeP2P) OnBlockAddition(block *kernel.Block) {
	ctx, cancel := context.WithTimeout(context.Background(), n.cfg.P2P.ConnTimeout)
	defer cancel()

	// notify all peers about the new block added
	if err := n.pubsub.NotifyBlockHeaderAdded(ctx, *block.Header); err != nil {
		n.logger.Errorf("error notifying block %x: %s", block.Hash, err)
	}
}
