package network

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/yago-123/chainnet/pkg/util"

	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/encoding"
	cerror "github.com/yago-123/chainnet/pkg/errs"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/observer"
)

type HTTPRouter struct {
	r          *httprouter.Router
	apiEncoder encoding.Encoding

	explorer   *explorer.ChainExplorer
	netSubject observer.NetSubject

	isActive bool
	srv      *http.Server

	logger *logrus.Logger
	cfg    *config.Config
}

const (
	NumberRetrievalsForConsideringActive = 1
	MaxNumberRetrievals                  = 100
)

func NewHTTPRouter(
	cfg *config.Config,
	explorer *explorer.ChainExplorer,
	netSubject observer.NetSubject,
) *HTTPRouter {
	router := &HTTPRouter{
		r: httprouter.New(),
		// by default the API encoder must be JSON due to OpenAPI spec generation.
		apiEncoder: encoding.NewJSONEncoder(),
		explorer:   explorer,
		netSubject: netSubject,
		logger:     cfg.Logger,
		cfg:        cfg,
	}

	router.r.GET(RouterV1BetaLatestChain, router.getLatestChain)
	router.r.GET(fmt.Sprintf(RouterV1BetaRetrieveAddressTxs, ":address"), router.listTransactions)
	router.r.GET(fmt.Sprintf(RouterV1BetaAddressIsActive, ":address"), router.checkAddressIsActive)
	router.r.GET(fmt.Sprintf(RouterV1BetaRetrieveAddressUTXOs, ":address"), router.listUTXOs)
	router.r.POST(RouterV1BetaSendTx, router.receiveTransaction)
	router.r.GET(fmt.Sprintf(RouterV1BetaTransactionByID, ":tx_id"), router.getTransactionByID)
	router.r.GET(fmt.Sprintf(RouterV1BetaBlockByHash, ":hash"), router.getBlock)
	router.r.GET(fmt.Sprintf(RouterV1BetaHeaderByHeight, ":height"), router.getHeader)
	router.r.GET(RouterV1BetaHeaders, router.listHeaders)

	return router
}

func (router *HTTPRouter) getLatestChain(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	block, err := router.explorer.GetLastBlock()
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve latest chain tip: %s", err.Error()), err)
		return
	}

	data := []byte(fmt.Sprintf(
		`{"height":%d,"hash":"%s","timestamp":%d}`,
		block.Header.Height,
		hex.EncodeToString(block.Hash),
		block.Header.Timestamp,
	))

	router.writeResponse(w, data)
}

func (router *HTTPRouter) Start() error {
	if router.isActive {
		return nil
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", router.cfg.P2P.RouterPort),
		Handler:      router.r,
		ReadTimeout:  router.cfg.P2P.ReadTimeout,
		WriteTimeout: router.cfg.P2P.WriteTimeout,
		IdleTimeout:  router.cfg.P2P.ConnTimeout,
	}

	router.srv = srv
	router.isActive = true

	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			router.logger.Infof("HTTP API server stopped successfully")
		}

		if err != nil {
			router.logger.Errorf("failed to start HTTP server: %v", err)
		}
	}()

	return nil
}

func (router *HTTPRouter) Stop() error {
	if !router.isActive {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ServerAPIShutdownTimeout)
	defer cancel()

	if err := router.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	router.isActive = false
	return nil
}

// listTransactions receive a wallet address and retrieve the unspent txs
func (router *HTTPRouter) listTransactions(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	addr, err := decodeAddress(address)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid address: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	txs, err := router.explorer.FindAllTransactions(addr, MaxNumberRetrievals)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to retrieve transactions: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	txsEncoded, err := router.apiEncoder.SerializeTransactions(txs)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode transactions: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	router.writeResponse(w, txsEncoded)
}

func (router *HTTPRouter) checkAddressIsActive(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	addr, err := decodeAddress(address)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid address: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	txs, err := router.explorer.FindAllTransactions(addr, NumberRetrievalsForConsideringActive)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to retrieve transactions: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	// check if there is any transaction for the address
	active, err := router.apiEncoder.SerializeBool(len(txs) > 0)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode active status: %s", err.Error()), http.StatusBadRequest, err)
	}

	router.writeResponse(w, active)
}

// listUTXOs receive a wallet address and retrieve the corresponding UTXOs
func (router *HTTPRouter) listUTXOs(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	addr, err := decodeAddress(address)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid address: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	utxos, err := router.explorer.FindUnspentOutputs(addr, MaxNumberRetrievals)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to retrieve UTXOs: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	utxosEncoded, err := router.apiEncoder.SerializeUTXOs(utxos)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode UTXOs: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	router.writeResponse(w, utxosEncoded)
}

// receiveTransaction receives txs sent by the wallets in order to be appended into the chain mempool via POST request
func (router *HTTPRouter) receiveTransaction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to read payload: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	tx, err := router.apiEncoder.DeserializeTransaction(data)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to decode transaction: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	if errObserver := router.netSubject.NotifyUnconfirmedTxReceived(*tx); errObserver != nil {
		router.handleError(w, fmt.Sprintf("Failed to append transaction: %s", errObserver.Error()), http.StatusBadRequest, errObserver)
	}
}

func (router *HTTPRouter) getTransactionByID(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	txID, err := decodeHash(ps.ByName("tx_id"))
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid transaction ID: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	tx, err := router.explorer.GetTransactionByID(txID)
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve transaction: %s", err.Error()), err)
		return
	}

	txEncoded, err := router.apiEncoder.SerializeTransaction(*tx)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode transaction: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, txEncoded)
}

func (router *HTTPRouter) getLatestBlock(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	block, err := router.explorer.GetLastBlock()
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve latest block: %s", err.Error()), err)
		return
	}

	blockEncoded, err := router.apiEncoder.SerializeBlock(*block)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode latest block: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, blockEncoded)
}

func (router *HTTPRouter) getBlock(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("hash") == "latest" {
		router.getLatestBlock(w, r, ps)
		return
	}

	router.getBlockByHash(w, r, ps)
}

func (router *HTTPRouter) getBlockByHash(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	blockHash, err := decodeHash(ps.ByName("hash"))
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid block hash: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	block, err := router.explorer.GetBlockByHash(blockHash)
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve block: %s", err.Error()), err)
		return
	}

	blockEncoded, err := router.apiEncoder.SerializeBlock(*block)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode block: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, blockEncoded)
}

func (router *HTTPRouter) getLatestHeader(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	header, err := router.explorer.GetLastHeader()
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve latest header: %s", err.Error()), err)
		return
	}

	headerEncoded, err := router.apiEncoder.SerializeHeader(*header)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode latest header: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, headerEncoded)
}

func (router *HTTPRouter) getHeader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("height") == "latest" {
		router.getLatestHeader(w, r, ps)
		return
	}

	router.getHeaderByHeight(w, r, ps)
}

func (router *HTTPRouter) getHeaderByHeight(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	height, err := strconv.ParseUint(ps.ByName("height"), 10, 0)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid header height: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	header, err := router.explorer.GetHeaderByHeight(uint(height))
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve header: %s", err.Error()), err)
		return
	}

	headerEncoded, err := router.apiEncoder.SerializeHeader(*header)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode header: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, headerEncoded)
}

func (router *HTTPRouter) listHeaders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	headers, err := router.explorer.GetAllHeaders()
	if err != nil {
		router.handleExplorerError(w, fmt.Sprintf("Failed to retrieve headers: %s", err.Error()), err)
		return
	}

	headers, err = filterHeaders(headers, r)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid headers query: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	headersEncoded, err := router.apiEncoder.SerializeHeaders(headers)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode headers: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	router.writeResponse(w, headersEncoded)
}

func (router *HTTPRouter) writeResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set(ContentTypeHeader, "application/json")
	if _, err := w.Write(data); err != nil {
		router.logger.Errorf("failed to write response: %v", err)
	}
}

func (router *HTTPRouter) handleError(w http.ResponseWriter, msg string, code int, logErr error) {
	http.Error(w, msg, code)
	if logErr != nil {
		router.logger.Errorf("%s: %v", msg, logErr)
	}
}

func (router *HTTPRouter) handleExplorerError(w http.ResponseWriter, msg string, err error) {
	if errors.Is(err, cerror.ErrStorageElementNotFound) {
		router.handleError(w, msg, http.StatusNotFound, err)
		return
	}

	router.handleError(w, msg, http.StatusInternalServerError, err)
}

func decodeAddress(address string) (string, error) {
	if !util.IsValidAddress([]byte(address)) {
		return "", fmt.Errorf("error validating address")
	}
	return string(base58.Decode(address)), nil
}

func decodeHash(hash string) ([]byte, error) {
	decoded, err := hex.DecodeString(hash)
	if err != nil {
		return []byte{}, fmt.Errorf("error decoding hash: %w", err)
	}

	if !util.IsValidHash(decoded) {
		return []byte{}, fmt.Errorf("error validating hash")
	}

	return decoded, nil
}

func filterHeaders(headers []*kernel.BlockHeader, r *http.Request) ([]*kernel.BlockHeader, error) {
	order := r.URL.Query().Get("order")
	switch order {
	case "", "newest_first":
	case "oldest_first":
		for left, right := 0, len(headers)-1; left < right; left, right = left+1, right-1 {
			headers[left], headers[right] = headers[right], headers[left]
		}
	default:
		return nil, fmt.Errorf("unsupported order %q", order)
	}

	limitRaw := r.URL.Query().Get("limit")
	if limitRaw == "" {
		return headers, nil
	}

	limit, err := strconv.ParseUint(limitRaw, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid limit: %w", err)
	}

	if limit == 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	if limit > uint64(MaxNumberRetrievals) {
		return nil, fmt.Errorf("limit must be less than or equal to %d", MaxNumberRetrievals)
	}

	limitInt := int(limit)
	if limitInt > len(headers) {
		return headers, nil
	}

	return headers[:limitInt], nil
}
