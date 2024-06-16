package main

import (
	blockchain "chainnet/pkg/chain"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func NewHTTPRouter(bc *blockchain.Blockchain) *httprouter.Router {
	router := httprouter.New()

	router.GET("/address/:address/transactions", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		listTransactions(w, r, ps, bc)
	})
	router.GET("/address/:address/utxos", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		listUTXOs(w, r, ps, bc)
	})
	router.GET("/address/:address/balance", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		getAddressBalance(w, r, ps, bc)
	})

	return router
}

func listTransactions(w http.ResponseWriter, r *http.Request, ps httprouter.Params, bc *blockchain.Blockchain) {
	address := ps.ByName("address")

	// todo() replace this method with all transactions instead of only non-spent ones
	transactions, err := bc.FindUnspentTransactions(address)
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
	}
}

func listUTXOs(w http.ResponseWriter, r *http.Request, ps httprouter.Params, bc *blockchain.Blockchain) {
	address := ps.ByName("address")

	utxos, err := bc.FindUnspentTransactions(address)
	if err != nil {
		http.Error(w, "Failed to retrieve utxos", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(utxos)
	if err != nil {
		http.Error(w, "Failed to encode UTXOs", http.StatusInternalServerError)
	}
}

func getAddressBalance(w http.ResponseWriter, r *http.Request, ps httprouter.Params, bc *blockchain.Blockchain) {
	address := ps.ByName("address")

	balanceResponse, err := bc.CalculateAddressBalance(address)
	if err != nil {
		http.Error(w, "Failed to find unspent transactions", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(balanceResponse)
	if err != nil {
		http.Error(w, "Failed to encode balance", http.StatusInternalServerError)
	}
}
