package network

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/yago-123/chainnet/pkg/monitor"

	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	cerror "github.com/yago-123/chainnet/pkg/errs"

	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"

	"github.com/yago-123/chainnet/pkg/mempool"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/network/discovery"
	"github.com/yago-123/chainnet/pkg/network/events"
	"github.com/yago-123/chainnet/pkg/network/pubsub"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/util"

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
	AskSpecificTxProtocol    = "/askSpecificTx/0.1.0"
	AskAllHeaders            = "/askAllHeaders/0.1.0"

	ServerAPIShutdownTimeout = 10 * time.Second
)

type nodeP2PHandler struct {
	logger          *logrus.Logger
	encoder         encoding.Encoding
	explorer        *explorer.ChainExplorer
	mempoolExplorer *mempool.MemPoolExplorer

	netSubject observer.NetSubject

	cfg *config.Config
}

func newNodeP2PHandler(
	cfg *config.Config,
	encoder encoding.Encoding,
	explorer *explorer.ChainExplorer,
	mempoolExplorer *mempool.MemPoolExplorer,
	netSubject observer.NetSubject,
) *nodeP2PHandler {
	return &nodeP2PHandler{
		logger:          cfg.Logger,
		encoder:         encoder,
		explorer:        explorer,
		mempoolExplorer: mempoolExplorer,
		netSubject:      netSubject,
		cfg:             cfg,
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
		if errors.Is(err, cerror.ErrStorageElementNotFound) {
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

	if valid := util.IsValidHash(hash); !valid {
		h.logger.Errorf("invalid hash %x received from stream %s", hash, stream.ID())
		return
	}

	// retrieve block from explorer
	block, err := h.explorer.GetBlockByHash(hash)
	if err != nil {
		if errors.Is(err, cerror.ErrStorageElementNotFound) {
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

func (h *nodeP2PHandler) handleAskSpecificTx(stream network.Stream) {
	// open stream with timeout
	timeoutStream := AddTimeoutToStream(stream, h.cfg)
	defer timeoutStream.Close()

	// read hash of transaction that is being requested
	txID, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		h.logger.Errorf("error reading transaction hash from stream %s: %s", stream.ID(), err)
		return
	}

	if valid := util.IsValidHash(txID); !valid {
		h.logger.Errorf("invalid hash %x received from stream %s", txID, stream.ID())
		return
	}

	// retrieve transaction from explorer
	tx, err := h.mempoolExplorer.RetrieveTx(string(txID))
	if err != nil {
		h.logger.Errorf("unable to retrieve transaction for stream %s: transaction %x not found", stream.ID(), txID)
		return
	}

	// encode transaction
	data, err := h.encoder.SerializeTransaction(*tx)
	if err != nil {
		h.logger.Errorf("error serializing transaction with hash %x for stream %s: %s", txID, stream.ID(), err)
		return
	}

	// send transaction encoded to the peer
	_, err = timeoutStream.WriteWithTimeout(data)
	if err != nil {
		h.logger.Errorf("error writing transaction with hash %x to stream %s: %s", txID, stream.ID(), err)
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
		if errors.Is(err, cerror.ErrStorageElementNotFound) {
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
	explorer *explorer.ChainExplorer

	router *HTTPRouter

	bandwithCounter *metrics.BandwidthCounter

	// bufferSize represents size of buffer for reading over the network
	bufferSize uint

	logger *logrus.Logger
}

func NewNodeP2P(
	ctx context.Context,
	cfg *config.Config,
	netSubject observer.NetSubject,
	encoder encoding.Encoding,
	explorer *explorer.ChainExplorer,
	mempoolExplorer *mempool.MemPoolExplorer,
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
	if cfg.P2P.IdentityPath != "" {
		privKeyBytes, errKey := util_crypto.ReadECDSAPemToPrivateKeyDerBytes(cfg.P2P.IdentityPath)
		if errKey != nil {
			return nil, fmt.Errorf("error reading private key: %w", errKey)
		}

		priv, errKey := util_crypto.ConvertDERBytesToECDSAPriv(privKeyBytes)
		if errKey != nil {
			return nil, fmt.Errorf("error converting private key: %w", errKey)
		}

		p2pPrivKey, _, errKey := p2pCrypto.ECDSAKeyPairFromKey(priv)
		if errKey != nil {
			return nil, fmt.Errorf("error creating p2p key pair: %w", errKey)
		}

		// add peer identity to options
		options = append(options, libp2p.Identity(p2pPrivKey))
	}

	// store bandwith metrics in counter for extracting metrics in RegisterMetrics
	bandwithCounter := metrics.NewBandwidthCounter()
	options = append(options, libp2p.BandwidthReporter(bandwithCounter))

	// add separate libp2p core metrics to a separate Prometheus registry
	if cfg.Prometheus.Enabled {
		registry := prometheus.NewRegistry()
		http.Handle(cfg.Prometheus.Path, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		go func() {
			cfg.Logger.Infof("exposing libp2p Prometheus metrics in http://localhost:%d%s", cfg.Prometheus.PortLibp2p, cfg.Prometheus.Path)
			if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Prometheus.PortLibp2p), nil); err != nil {
				cfg.Logger.Errorf("failed to start metrics server: %v", err)
			}
		}()

		options = append(options, libp2p.PrometheusRegisterer(registry))
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
	router := NewHTTPRouter(cfg, encoder, explorer, netSubject)

	// initialize handlers
	handler := newNodeP2PHandler(cfg, encoder, explorer, mempoolExplorer, netSubject)
	host.SetStreamHandler(AskLastHeaderProtocol, handler.handleAskLastHeader)
	host.SetStreamHandler(AskSpecificBlockProtocol, handler.handleAskSpecificBlock)
	host.SetStreamHandler(AskSpecificTxProtocol, handler.handleAskSpecificTx)
	host.SetStreamHandler(AskAllHeaders, handler.handleAskAllHeaders)

	return &NodeP2P{
		cfg:             cfg,
		host:            host,
		netSubject:      netSubject,
		ctx:             ctx,
		discoDHT:        discoDHT,
		discoMDNS:       discoMDNS,
		pubsub:          pubsub,
		encoder:         encoder,
		router:          router,
		explorer:        explorer,
		bandwithCounter: bandwithCounter,
		bufferSize:      cfg.P2P.BufferSize,
		logger:          cfg.Logger,
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

// ConnectToSeeds creates connection with the seed nodes
func (n *NodeP2P) ConnectToSeeds() error {
	return connectToSeeds(n.cfg, n.host)
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

func (n *NodeP2P) AskSpecificTx(ctx context.Context, peerID peer.ID, txID []byte) (*kernel.Transaction, error) {
	// open stream to peer with timeout
	timeoutStream, err := NewTimeoutStream(ctx, n.cfg, n.host, peerID, AskSpecificTxProtocol)
	if err != nil {
		return nil, err
	}
	defer timeoutStream.Close()

	// write transaction hash required to stream
	_, err = timeoutStream.WriteWithTimeout(txID)
	if err != nil {
		return nil, fmt.Errorf("error writing transaction hash %x to stream: %w", txID, err)
	}
	// close write side of the stream so the peer knows we are done writing
	err = timeoutStream.stream.CloseWrite()
	if err != nil {
		return nil, fmt.Errorf("error closing write side of the stream: %w", err)
	}

	// read and decode transaction retrieved
	data, err := timeoutStream.ReadWithTimeout()
	if err != nil {
		return nil, fmt.Errorf("error reading data from stream: %w", err)
	}

	return n.encoder.DeserializeTransaction(data)
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

// OnTxAddition is triggered when a new transaction is added into the MemPool
func (n *NodeP2P) OnTxAddition(tx *kernel.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), n.cfg.P2P.ConnTimeout)
	defer cancel()

	// notify all peers about the new transaction added
	if err := n.pubsub.NotifyTransactionAdded(ctx, *tx); err != nil {
		n.logger.Errorf("error notifying transaction %x: %s", tx.ID, err)
	}
}

func (n *NodeP2P) RegisterMetrics(register *prometheus.Registry) {
	monitor.NewMetric(register, monitor.Counter, "bandwidth_total_incoming_bytes", "Total incoming bandwidth in bytes",
		func() float64 {
			return float64(n.bandwithCounter.GetBandwidthTotals().TotalIn)
		},
	)

	monitor.NewMetric(register, monitor.Counter, "bandwidth_total_outgoing_bytes", "Total outgoing bandwidth in bytes",
		func() float64 {
			return float64(n.bandwithCounter.GetBandwidthTotals().TotalOut)
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "bandwidth_rate_incoming_bytes", "Incoming bandwidth rate in bytes per second",
		func() float64 {
			return float64(n.bandwithCounter.GetBandwidthTotals().RateIn)
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "bandwidth_rate_outgoing_bytes", "Outgoing bandwidth rate in bytes per second",
		func() float64 {
			return float64(n.bandwithCounter.GetBandwidthTotals().RateOut)
		},
	)

	monitor.NewMetricWithLabelsAsync(register, monitor.Gauge, "bandwidth_total_bytes_by_protocol", "Bandwidth total statistics by protocol",
		[]string{monitor.OperationLabel, monitor.ProtocolLabel},
		func(metricVec interface{}) {
			gaugeVec := metricVec.(*prometheus.GaugeVec)
			for {
				stats := n.bandwithCounter.GetBandwidthByProtocol()
				for protocol, stat := range stats {
					gaugeVec.WithLabelValues("in", string(protocol)).Set(float64(stat.TotalIn))
					gaugeVec.WithLabelValues("out", string(protocol)).Set(float64(stat.TotalOut))
				}
				time.Sleep(n.cfg.Prometheus.UpdateInterval)
			}
		})

	monitor.NewMetricWithLabelsAsync(register, monitor.Gauge, "bandwidth_rate_bytes_by_protocol", "Bandwidth rate statistics by protocol",
		[]string{monitor.OperationLabel, monitor.ProtocolLabel},
		func(metricVec interface{}) {
			gaugeVec := metricVec.(*prometheus.GaugeVec)
			for {
				stats := n.bandwithCounter.GetBandwidthByProtocol()
				for protocol, stat := range stats {
					gaugeVec.WithLabelValues("in", string(protocol)).Set(stat.RateIn)
					gaugeVec.WithLabelValues("out", string(protocol)).Set(stat.RateOut)
				}
				time.Sleep(n.cfg.Prometheus.UpdateInterval)
			}
		})

	monitor.NewMetricWithLabelsAsync(register, monitor.Gauge, "bandwidth_total_bytes_by_peer", "Bandwidth total statistics by peer",
		[]string{monitor.OperationLabel, monitor.PeerLabel},
		func(metricVec interface{}) {
			gaugeVec := metricVec.(*prometheus.GaugeVec)
			go func() {
				for {
					stats := n.bandwithCounter.GetBandwidthByPeer()
					for peerID, stat := range stats {
						gaugeVec.WithLabelValues("in", peerID.String()).Set(float64(stat.TotalIn))
						gaugeVec.WithLabelValues("out", peerID.String()).Set(float64(stat.TotalOut))
					}

					time.Sleep(n.cfg.Prometheus.UpdateInterval)
				}
			}()
		})

	monitor.NewMetricWithLabelsAsync(register, monitor.Gauge, "bandwidth_rate_bytes_by_peer", "Bandwidth rate statistics by peer",
		[]string{monitor.OperationLabel, monitor.PeerLabel},
		func(metricVec interface{}) {
			gaugeVec := metricVec.(*prometheus.GaugeVec)
			go func() {
				for {
					stats := n.bandwithCounter.GetBandwidthByPeer()
					for peerID, stat := range stats {
						gaugeVec.WithLabelValues("in", peerID.String()).Set(float64(stat.RateIn))
						gaugeVec.WithLabelValues("out", peerID.String()).Set(float64(stat.RateOut))
					}

					time.Sleep(n.cfg.Prometheus.UpdateInterval)
				}
			}()
		})
}
