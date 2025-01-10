package network

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
	"github.com/yago-123/chainnet/pkg/network/discovery"
)

const RequestTimeout = 10 * time.Second

type WalletNetwork interface {
	GetWalletUTXOS(ctx context.Context, address []byte) ([]*kernel.UTXO, error)
	GetWalletTxs(ctx context.Context, address []byte) ([]*kernel.Transaction, error)
	SendTransaction(ctx context.Context, tx kernel.Transaction) error
}

type WalletHTTPConn struct {
	cfg *config.Config

	host host.Host

	// disco is in charge of setting up the logic for node discovery
	disco discovery.Discovery
	// encoder contains the communication data serialization between peers
	encoder encoding.Encoding

	baseurl string

	logger *logrus.Logger
}

func NewWalletHTTPConn(
	cfg *config.Config,
	encoder encoding.Encoding,
) (WalletNetwork, error) {
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
	
	// initialize discovery module
	disco, err := discovery.NewDHTDiscovery(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT discovery module: %w", err)
	}

	baseURL := net.JoinHostPort(cfg.Wallet.ServerAddress, fmt.Sprintf("%d", cfg.Wallet.ServerPort))

	return &WalletHTTPConn{
		cfg:     cfg,
		host:    host,
		disco:   disco,
		encoder: encoder,
		baseurl: baseURL,
		logger:  cfg.Logger,
	}, nil
}

// GetWalletUTXOS returns the UTXOs for a given address
func (n *WalletHTTPConn) GetWalletUTXOS(ctx context.Context, address []byte) ([]*kernel.UTXO, error) {
	return fetchAndDecode(
		ctx,
		n.baseurl,
		address,
		RouterRetrieveAddressUTXOs,
		n.getRequest,
		n.encoder.DeserializeUTXOs,
	)
}

// GetWalletTxs returns the transactions for a given address
func (n *WalletHTTPConn) GetWalletTxs(ctx context.Context, address []byte) ([]*kernel.Transaction, error) {
	return fetchAndDecode(
		ctx,
		n.baseurl,
		address,
		RouterRetrieveAddressTxs,
		n.getRequest,
		n.encoder.DeserializeTransactions,
	)
}

func (n *WalletHTTPConn) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
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

// helper function to send HTTP request and decode response with URL in error messages
func fetchAndDecode[T any](
	ctx context.Context,
	baseurl string,
	address []byte,
	routeFormat string,
	getRequest func(context.Context, string) (*http.Response, error),
	decodeFunc func([]byte) ([]T, error),
) ([]T, error) {
	// construct the URL
	url := fmt.Sprintf(
		"http://%s%s",
		baseurl,
		fmt.Sprintf(routeFormat, base58.Encode(address)),
	)

	// send GET request
	resp, err := getRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to GET from URL %s for address %s: %w", url, base58.Encode(address), err)
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from URL %s for address %s: %w", url, base58.Encode(address), err)
	}

	// decode response
	data, err := decodeFunc(body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response from URL %s for address %s: %w", url, base58.Encode(address), err)
	}

	return data, nil
}

func (n *WalletHTTPConn) getRequest(ctx context.Context, url string) (*http.Response, error) {
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

func (n *WalletHTTPConn) postRequest(ctx context.Context, url string, payload io.Reader) (*http.Response, error) {
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
