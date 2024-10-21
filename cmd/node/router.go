package main

import (
	"encoding/json"
	"net/http"

	"github.com/btcsuite/btcutil/base58"

	blockchain "github.com/yago-123/chainnet/pkg/chain/explorer"

	"github.com/julienschmidt/httprouter"
)

func NewHTTPRouter(explorer *blockchain.Explorer) *httprouter.Router {
	router := httprouter.New()

	router.GET("/address/:address/transactions", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		listTransactions(w, r, ps, explorer)
	})
	router.GET("/address/:address/utxos", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		listUTXOs(w, r, ps, explorer)
	})
	router.GET("/address/:address/balance", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		getAddressBalance(w, r, ps, explorer)
	})

	return router
}

func listTransactions(w http.ResponseWriter, _ *http.Request, ps httprouter.Params, explorer *blockchain.Explorer) {
	address := ps.ByName("address")

	// todo() replace this method with all transactions instead of only non-spent ones
	transactions, err := explorer.FindUnspentTransactions(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(transactions) //nolint:musttag // not sure which encoding will use in the future
	if err != nil {
		http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
	}
}

func listUTXOs(w http.ResponseWriter, _ *http.Request, ps httprouter.Params, explorer *blockchain.Explorer) {
	address := ps.ByName("address")

	utxos, err := explorer.FindUnspentTransactions(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to retrieve utxos", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(utxos) //nolint:musttag // not sure which encoding will use in the future
	if err != nil {
		http.Error(w, "Failed to encode UTXOs", http.StatusInternalServerError)
	}
}

func getAddressBalance(w http.ResponseWriter, _ *http.Request, ps httprouter.Params, explorer *blockchain.Explorer) {
	address := ps.ByName("address")

	balanceResponse, err := explorer.CalculateAddressBalance(string(base58.Decode(address)))
	if err != nil {
		http.Error(w, "Failed to find unspent transactions", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(balanceResponse)
	if err != nil {
		http.Error(w, "Failed to encode balance", http.StatusInternalServerError)
	}
}
