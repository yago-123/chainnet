package v1beta //nolint:testpackage // keep tests in package to exercise unexported helpers during SDK iteration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientWalletEndpoints(t *testing.T) { //nolint:gocognit // endpoint smoke test is intentionally linear
	utxos := []UTXO{
		{
			Txid: "74782d6964",
			Vout: 1,
			Output: TxOutput{
				Amount:       10,
				ScriptPubKey: "script",
				PubKey:       "pub-key",
			},
		},
	}
	txs := []Transaction{
		{
			Id: "74782d6964",
			Vin: []TxInput{
				{Txid: "74782d6964", Vout: 1, ScriptSig: "script-sig", PubKey: "pub-key"},
			},
			Vout: []TxOutput{
				{Amount: 10, ScriptPubKey: "script", PubKey: "pub-key"},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/utxos", writeJSON(t, utxos))
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/transactions", writeJSON(t, txs))
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/activity", writeJSON(t, true))
	mux.HandleFunc("/api/v1beta/transactions", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("content type = %q, want %q", got, want)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var tx Transaction
		if err := json.Unmarshal(body, &tx); err != nil {
			t.Fatal(err)
		}
		if tx.Id != "74782d6964" {
			t.Fatalf("tx ID = %q, want %q", tx.Id, "74782d6964")
		}

		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/v1beta/transactions/74782d6964", writeJSON(t, txs[0]))

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	gotUTXOs, err := client.GetAddressUTXOs(context.Background(), []byte("pub-key"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gotUTXOs) != 1 || gotUTXOs[0].Vout != 1 {
		t.Fatalf("unexpected UTXOs: %#v", gotUTXOs)
	}

	gotTxs, err := client.GetAddressTransactions(context.Background(), []byte("pub-key"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gotTxs) != 1 || gotTxs[0].Id != "74782d6964" {
		t.Fatalf("unexpected transactions: %#v", gotTxs)
	}

	active, err := client.AddressIsActive(context.Background(), []byte("pub-key"))
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("expected address to be active")
	}

	if sendErr := client.SendTransaction(context.Background(), txs[0]); sendErr != nil {
		t.Fatal(sendErr)
	}

	tx, err := client.GetTransactionByID(context.Background(), []byte("tx-id"))
	if err != nil {
		t.Fatal(err)
	}
	if tx.Id != "74782d6964" {
		t.Fatalf("tx ID = %q, want %q", tx.Id, "74782d6964")
	}
}

func TestClientChainEndpoints(t *testing.T) { //nolint:gocognit // endpoint smoke test is intentionally linear
	header := BlockHeader{
		Version:       "7631",
		PrevBlockHash: "70726576696f75732d626c6f636b2d68617368",
		MerkleRoot:    "6d65726b6c652d726f6f74",
		Height:        7,
		Timestamp:     123,
		Target:        1,
		Nonce:         42,
	}
	block := Block{
		Header:       header,
		Transactions: []Transaction{},
		Hash:         "626c6f636b2d68617368",
	}
	tip := ChainTip{
		Height:    7,
		Hash:      "626c6f636b2d68617368",
		Timestamp: 123,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1beta/chain/latest", writeJSON(t, tip))
	mux.HandleFunc("/api/v1beta/blocks/latest", writeJSON(t, block))
	mux.HandleFunc("/api/v1beta/blocks/626c6f636b2d68617368", writeJSON(t, block))
	mux.HandleFunc("/api/v1beta/headers/latest", writeJSON(t, header))
	mux.HandleFunc("/api/v1beta/headers/7", writeJSON(t, header))
	mux.HandleFunc("/api/v1beta/headers", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("limit"), "1"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("order"), string(HeaderOrderOldestFirst); got != want {
			t.Fatalf("order = %q, want %q", got, want)
		}
		writeJSON(t, []BlockHeader{header})(w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	gotTip, err := client.GetLatestChain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotTip.Height != 7 || gotTip.Hash != "626c6f636b2d68617368" {
		t.Fatalf("unexpected chain tip: %#v", gotTip)
	}

	gotBlock, err := client.GetLatestBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotBlock.Hash != "626c6f636b2d68617368" {
		t.Fatalf("latest block hash = %q, want %q", gotBlock.Hash, "626c6f636b2d68617368")
	}

	gotBlock, err = client.GetBlockByHash(context.Background(), []byte("block-hash"))
	if err != nil {
		t.Fatal(err)
	}
	if gotBlock.Hash != "626c6f636b2d68617368" {
		t.Fatalf("block hash = %q, want %q", gotBlock.Hash, "626c6f636b2d68617368")
	}

	gotHeader, err := client.GetLatestHeader(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotHeader.Height != 7 {
		t.Fatalf("latest header height = %d, want 7", gotHeader.Height)
	}

	gotHeader, err = client.GetHeaderByHeight(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if gotHeader.Height != 7 {
		t.Fatalf("header height = %d, want 7", gotHeader.Height)
	}

	limit := 1
	order := HeaderOrderOldestFirst
	gotHeaders, err := client.GetHeaders(context.Background(), &GetHeadersOptions{Limit: &limit, Order: &order})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotHeaders) != 1 || gotHeaders[0].Height != 7 {
		t.Fatalf("unexpected headers: %#v", gotHeaders)
	}
}

func TestClientUnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1beta/blocks/latest", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetLatestBlock(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeJSON(t *testing.T, value any) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(value); err != nil {
			t.Fatal(err)
		}
	}
}
