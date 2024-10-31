package net

import (
	"context"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/observer"
	"io"
	"net/http"
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

	router.r.GET(fmt.Sprintf(RouterRetrieveAddressTxs, ":address"), func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router.listTransactions(w, r, ps)
	})
	router.r.GET(fmt.Sprintf(RouterRetrieveAddressUTXOs, ":address"), func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router.listUTXOs(w, r, ps)
	})
	router.r.POST(RouterSendTx, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		router.receiveTransaction(w, r)
	})

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
			router.logger.Errorf("Failed to start HTTP server: %v", err)
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

	txs, err := router.explorer.FindUnspentTransactions(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
	}

	txsEncoded, err := router.encoder.SerializeTransactions(txs)
	if err != nil {
		http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
	}

	w.Header().Set(ContentTypeHeader, getContentTypeFrom(router.encoder))
	if _, err = w.Write(txsEncoded); err != nil {
		router.logger.Errorf("Failed to write response: %v", err)
	}
}

// listUTXOs receive a wallet address and retrieve the corresponding UTXOs
func (router *HTTPRouter) listUTXOs(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")

	utxos, err := router.explorer.FindUnspentOutputs(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve utxos", http.StatusInternalServerError)
	}

	utxosEncoded, err := router.encoder.SerializeUTXOs(utxos)
	if err != nil {
		http.Error(w, "Failed to encode utxos", http.StatusInternalServerError)
	}

	w.Header().Set(ContentTypeHeader, getContentTypeFrom(router.encoder))
	if _, err = w.Write(utxosEncoded); err != nil {
		router.logger.Errorf("Failed to write response: %v", err)
	}
}

// receiveTransaction receives txs sent by the wallets in order to be appended into the chain mempool via POST request
func (router *HTTPRouter) receiveTransaction(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	tx, err := router.encoder.DeserializeTransaction(data)
	if err != nil {
		http.Error(w, "Failed to decode transaction", http.StatusInternalServerError)
	}

	if errObserver := router.netSubject.NotifyUnconfirmedTxReceived(*tx); errObserver != nil {
		http.Error(w, fmt.Sprintf("Failed to append transaction: %v", errObserver), http.StatusExpectationFailed)
	}
}
