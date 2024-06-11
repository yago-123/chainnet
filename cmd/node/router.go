package main

import (
	"chainnet/pkg/block"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func NewHTTPRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/transactions", listTransactions)
	router.GET("/utxos", listUTXOs)
	router.GET("/blocks", listBlocks)
	router.GET("/wallet/:wallet/balance", getWalletBalance)

	return router
}

func listTransactions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	transactions := []block.Transaction{}
	err := json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
	}
}

func listUTXOs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	utxos := []block.Transaction{}
	err := json.NewEncoder(w).Encode(utxos)
	if err != nil {
		http.Error(w, "Failed to encode UTXOs", http.StatusInternalServerError)
	}
}

func listBlocks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	blocks := []block.Block{}
	err := json.NewEncoder(w).Encode(blocks)
	if err != nil {
		http.Error(w, "Failed to encode blocks", http.StatusInternalServerError)
	}
}

func getWalletBalance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wallet := ps.ByName("wallet")

	// construct the URL using the base URL and the wallet address
	url := fmt.Sprintf("wallet/%s/balance", wallet)

	// make an HTTP GET request to the URL to retrieve the balance
	response, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to retrieve balance", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// decode the response body and write it to the response writer
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Failed to retrieve balance", response.StatusCode)
		return
	}
	balanceResponse := uint(0)
	if err := json.NewDecoder(response.Body).Decode(&balanceResponse); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// write the balance response to the response writer
	err = json.NewEncoder(w).Encode(balanceResponse)
	if err != nil {
		http.Error(w, "Failed to encode balance", http.StatusInternalServerError)
	}
}
