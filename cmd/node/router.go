package node

import (
	"chainnet/pkg/block"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/transactions", listTransactions)
	router.GET("/utxos", listUTXOs)
	router.GET("/blocks", listBlocks)
	router.GET("/wallet/:wallet/balance", getWalletBalance)

	return router
}

func listTransactions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	transactions := []block.Transaction{}
	json.NewEncoder(w).Encode(transactions)
}

func listUTXOs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	utxos := []block.Transaction{}
	json.NewEncoder(w).Encode(utxos)
}

func listBlocks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	blocks := []block.Block{}
	json.NewEncoder(w).Encode(blocks)
}

func getWalletBalance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wallet := ps.ByName("wallet")

	// Construct the URL using the base URL and the wallet address
	url := fmt.Sprintf("wallet/%s/balance", wallet)

	// Make an HTTP GET request to the URL to retrieve the balance
	response, err := http.Get(url)
	if err != nil {
		// Handle error
		http.Error(w, "Failed to retrieve balance", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// Decode the response body and write it to the response writer
	if response.StatusCode != http.StatusOK {
		// Handle non-200 status code
		http.Error(w, "Failed to retrieve balance", response.StatusCode)
		return
	}
	balanceResponse := uint(0)
	if err := json.NewDecoder(response.Body).Decode(&balanceResponse); err != nil {
		// Handle decoding error
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Write the balance response to the response writer
	json.NewEncoder(w).Encode(balanceResponse)
}
