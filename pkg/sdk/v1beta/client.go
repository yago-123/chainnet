package v1beta

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/sdk/v1beta/generated"
)

const apiBasePath = "/api/v1beta"

type Client struct {
	client *generated.ClientWithResponses
}

func NewClient(baseurl string, httpClient *http.Client) (*Client, error) {
	opts := []generated.ClientOption{}
	if httpClient != nil {
		opts = append(opts, generated.WithHTTPClient(httpClient))
	}

	client, err := generated.NewClientWithResponses(normalizeServerURL(baseurl), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create generated v1beta client: %w", err)
	}

	return &Client{client: client}, nil
}

func NewClientFromConfig(cfg *config.Config, httpClient *http.Client) (*Client, error) {
	baseURL := net.JoinHostPort(cfg.Wallet.ServerAddress, fmt.Sprintf("%d", cfg.Wallet.ServerPort))
	return NewClient(baseURL, httpClient)
}

func (c *Client) GetAddressUTXOs(ctx context.Context, address []byte) ([]*kernel.UTXO, error) {
	resp, err := c.client.GetAddressUTXOsWithResponse(ctx, base58.Encode(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get address UTXOs: %w", err)
	}

	utxos, err := expectOK("get address UTXOs", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return utxosFromGenerated(utxos)
}

func (c *Client) GetAddressTransactions(ctx context.Context, address []byte) ([]*kernel.Transaction, error) {
	resp, err := c.client.GetAddressTransactionsWithResponse(ctx, base58.Encode(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get address transactions: %w", err)
	}

	transactions, err := expectOK("get address transactions", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return transactionsFromGenerated(transactions)
}

func (c *Client) AddressIsActive(ctx context.Context, address []byte) (bool, error) {
	resp, err := c.client.GetAddressActivityWithResponse(ctx, base58.Encode(address))
	if err != nil {
		return false, fmt.Errorf("failed to get address activity: %w", err)
	}

	return expectOK("get address activity", resp.StatusCode(), resp.Body, resp.JSON200)
}

func (c *Client) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	body, err := transactionToGenerated(tx)
	if err != nil {
		return err
	}

	resp, err := c.client.SubmitTransactionWithResponse(ctx, body)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return expectStatusOK("submit transaction", resp.StatusCode(), resp.Body)
}

func (c *Client) GetTransactionByID(ctx context.Context, txID []byte) (*kernel.Transaction, error) {
	resp, err := c.client.GetTransactionWithResponse(ctx, encodeHash(txID))
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	tx, err := expectOK("get transaction", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return transactionFromGenerated(tx)
}

func (c *Client) GetLatestBlock(ctx context.Context) (*kernel.Block, error) {
	resp, err := c.client.GetLatestBlockWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	block, err := expectOK("get latest block", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return blockFromGenerated(block)
}

func (c *Client) GetBlockByHash(ctx context.Context, hash []byte) (*kernel.Block, error) {
	resp, err := c.client.GetBlockByHashWithResponse(ctx, encodeHash(hash))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	block, err := expectOK("get block by hash", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return blockFromGenerated(block)
}

func (c *Client) GetLatestHeader(ctx context.Context) (*kernel.BlockHeader, error) {
	resp, err := c.client.GetLatestHeaderWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest header: %w", err)
	}

	header, err := expectOK("get latest header", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return headerFromGenerated(header)
}

func (c *Client) GetHeaderByHeight(ctx context.Context, height uint) (*kernel.BlockHeader, error) {
	convertedHeight, err := uintToInt("height", height)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.GetHeaderByHeightWithResponse(ctx, convertedHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get header by height: %w", err)
	}

	header, err := expectOK("get header by height", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return headerFromGenerated(header)
}

func (c *Client) GetHeaders(ctx context.Context) ([]*kernel.BlockHeader, error) {
	resp, err := c.client.GetHeadersWithResponse(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	headers, err := expectOK("get headers", resp.StatusCode(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}

	return headersFromGenerated(headers)
}

func normalizeServerURL(baseurl string) string {
	if !strings.HasPrefix(baseurl, "http://") && !strings.HasPrefix(baseurl, "https://") {
		baseurl = "http://" + baseurl
	}

	baseurl = strings.TrimRight(baseurl, "/")
	if strings.HasSuffix(baseurl, apiBasePath) {
		return baseurl
	}

	return baseurl + apiBasePath
}

func expectOK[T any](operation string, statusCode int, body []byte, payload *T) (T, error) {
	var zero T
	if payload == nil {
		return zero, unexpectedStatus(operation, statusCode, body)
	}

	return *payload, nil
}

func expectStatusOK(operation string, statusCode int, body []byte) error {
	if statusCode != http.StatusOK {
		return unexpectedStatus(operation, statusCode, body)
	}

	return nil
}

func unexpectedStatus(operation string, statusCode int, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("%s failed with response code %d", operation, statusCode)
	}

	return fmt.Errorf("%s failed with response code %d, message: %s", operation, statusCode, body)
}
