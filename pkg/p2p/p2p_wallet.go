package p2p

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/btcsuite/btcutil/base58"

	"github.com/yago-123/chainnet/pkg/consensus/util"

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

const RequestTimeout = 10 * time.Second

type WalletP2P struct {
	cfg *config.Config

	host host.Host
	ctx  context.Context

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery
	// pubsub is in charge of setting up the logic for data propagation
	pubsub pubsub.PubSub
	// encoder contains the communication data serialization between peers
	encoder encoding.Encoding

	baseurl string

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
	topics := []string{pubsub.TxAddedPubSubTopic}
	pubsub, err := pubsub.NewGossipPubSub(ctx, cfg, host, encoder, observer.NewNetSubject(), topics, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	baseURL := net.JoinHostPort(cfg.Wallet.ServerAddress, fmt.Sprintf("%d", cfg.Wallet.ServerPort))

	return &WalletP2P{
		cfg:     cfg,
		host:    host,
		ctx:     ctx,
		disco:   disco,
		pubsub:  pubsub,
		encoder: encoder,
		baseurl: baseURL,
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

// ConnectToSeeds creates connection with the seed nodes
func (n *WalletP2P) ConnectToSeeds() error {
	return connectToSeeds(n.cfg, n.host)
}

// GetWalletUTXOS returns the UTXOs for a given address
func (n *WalletP2P) GetWalletUTXOS(address []byte) ([]*kernel.UTXO, error) {
	if !util.IsValidAddress(address) {
		return []*kernel.UTXO{}, fmt.Errorf("invalid address format")
	}

	url := fmt.Sprintf(
		"http://%s%s",
		n.baseurl,
		fmt.Sprintf(RouterAddressUTXOs, base58.Encode(address)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	// send GET request
	resp, err := getRequest(ctx, url)
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("failed to get UTXO response for address %s: %w", base58.Encode(address), err)
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("failed to read list of UTXO response for address %s: %w", base58.Encode(address), err)
	}

	// decode UTXOs
	utxos, err := n.encoder.DeserializeUTXOs(body)
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("failed to unmarshal UTXO response for address %s: %w", address, err)
	}

	return utxos, nil
}

func (n *WalletP2P) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	return n.pubsub.NotifyTransactionAdded(ctx, tx)
}

func getRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", url, err)
	}

	return resp, nil
}
