package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/btcsuite/btcutil/base58"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/net/discovery"
)

const RequestTimeout = 10 * time.Second

type WalletP2P struct {
	cfg *config.Config

	host host.Host

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery
	// encoder contains the communication data serialization between peers
	encoder encoding.Encoding

	baseurl string

	logger *logrus.Logger
}

func NewWalletP2P(
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

	baseURL := net.JoinHostPort(cfg.Wallet.ServerAddress, fmt.Sprintf("%d", cfg.Wallet.ServerPort))

	return &WalletP2P{
		cfg:     cfg,
		host:    host,
		disco:   disco,
		encoder: encoder,
		baseurl: baseURL,
		logger:  cfg.Logger,
	}, nil
}

// GetWalletUTXOS returns the UTXOs for a given address
func (n *WalletP2P) GetWalletUTXOS(ctx context.Context, address []byte) ([]*kernel.UTXO, error) {
	url := fmt.Sprintf(
		"http://%s%s",
		n.baseurl,
		fmt.Sprintf(RouterRetrieveAddressUTXOs, base58.Encode(address)),
	)

	// send GET request
	resp, err := n.getRequest(ctx, url)
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
	url := fmt.Sprintf(
		"http://%s%s",
		n.baseurl,
		RouterSendTx,
	)

	data, err := n.encoder.SerializeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// send request containing the transaction encoded
	resp, err := n.postRequest(ctx, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (n *WalletP2P) getRequest(ctx context.Context, url string) (*http.Response, error) {
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

func (n *WalletP2P) postRequest(ctx context.Context, url string, payload io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", url, err)
	}

	req.Header.Set(ContentTypeHeader, getContentTypeFrom(n.encoder))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for url %s: %w", url, err)
	}

	// handle response status code to reduce redundancy in the code
	if resp.StatusCode != http.StatusOK {
		responseMsg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("response code %d, message: %s", resp.StatusCode, responseMsg)
	}

	return resp, nil
}
