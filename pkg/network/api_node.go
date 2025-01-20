package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/yago-123/chainnet/pkg/util"

	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/observer"
)

type HTTPRouter struct {
	r       *httprouter.Router
	encoder encoding.Encoding

	explorer   *explorer.ChainExplorer
	netSubject observer.NetSubject

	isActive bool
	srv      *http.Server

	logger *logrus.Logger
	cfg    *config.Config
}

const (
	MaxNumberRetrievals = 100
)

func NewHTTPRouter(
	cfg *config.Config,
	encoder encoding.Encoding,
	explorer *explorer.ChainExplorer,
	netSubject observer.NetSubject,
) *HTTPRouter {
	router := &HTTPRouter{
		r:          httprouter.New(),
		encoder:    encoder,
		explorer:   explorer,
		netSubject: netSubject,
		logger:     cfg.Logger,
		cfg:        cfg,
	}

	router.r.GET(fmt.Sprintf(RouterRetrieveAddressTxs, ":address"), router.listTransactions)
	router.r.GET(fmt.Sprintf(RouterRetrieveAddressUTXOs, ":address"), router.listUTXOs)
	router.r.POST(RouterSendTx, router.receiveTransaction)

	return router
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

// todo(): right now there is 0 usage for this
// listTransactions receive a wallet address and retrieve the unspent txs
func (router *HTTPRouter) listTransactions(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	addr, err := decodeAddress(address)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Invalid address: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	txs, err := router.explorer.FindAllTransactions(addr)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to retrieve transactions: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	txsEncoded, err := router.encoder.SerializeTransactions(txs)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to encode transactions: %s", err.Error()), http.StatusBadRequest, err)
		return
	}

	router.writeResponse(w, txsEncoded)
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

	utxosEncoded, err := router.encoder.SerializeUTXOs(utxos)
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

	tx, err := router.encoder.DeserializeTransaction(data)
	if err != nil {
		router.handleError(w, fmt.Sprintf("Failed to decode transaction: %s", err.Error()), http.StatusInternalServerError, err)
		return
	}

	if errObserver := router.netSubject.NotifyUnconfirmedTxReceived(*tx); errObserver != nil {
		router.handleError(w, fmt.Sprintf("Failed to append transaction: %s", errObserver.Error()), http.StatusBadRequest, errObserver)
	}
}

func (router *HTTPRouter) writeResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set(ContentTypeHeader, getContentTypeFrom(router.encoder))
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

func decodeAddress(address string) (string, error) {
	if !util.IsValidAddress([]byte(address)) {
		return "", fmt.Errorf("error validating address")
	}
	return string(base58.Decode(address)), nil
}
