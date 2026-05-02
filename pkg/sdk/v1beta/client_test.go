package v1beta //nolint:testpackage // keep tests in package to exercise unexported helpers during SDK iteration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
)

func TestClientWalletEndpoints(t *testing.T) { //nolint:gocognit // endpoint smoke test is intentionally linear
	encoder := encoding.NewJSONEncoder()

	utxos := []*kernel.UTXO{
		{
			TxID:   []byte("tx-id"),
			OutIdx: 1,
			Output: kernel.TxOutput{
				Amount:       10,
				ScriptPubKey: "script",
				PubKey:       "pub-key",
			},
		},
	}
	txs := []*kernel.Transaction{
		kernel.NewTransaction(
			[]kernel.TxInput{kernel.NewInput([]byte("tx-id"), 1, "script-sig", "pub-key")},
			[]kernel.TxOutput{{Amount: 10, ScriptPubKey: "script", PubKey: "pub-key"}},
		),
	}
	txs[0].SetID([]byte("tx-id"))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/utxos", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeUTXOs(utxos)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/transactions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeTransactions(txs)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/addresses/5GBGnW23s6/activity", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeBool(true)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/transactions", func(_ http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("content type = %q, want %q", got, want)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := encoder.DeserializeTransaction(body)
		if err != nil {
			t.Fatal(err)
		}
		if string(tx.ID) != "tx-id" {
			t.Fatalf("tx ID = %q, want %q", tx.ID, "tx-id")
		}
	})
	mux.HandleFunc("/api/v1beta/transactions/74782d6964", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeTransaction(*txs[0])
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})

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
	if len(gotUTXOs) != 1 || gotUTXOs[0].OutIdx != 1 {
		t.Fatalf("unexpected UTXOs: %#v", gotUTXOs)
	}

	gotTxs, err := client.GetAddressTransactions(context.Background(), []byte("pub-key"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gotTxs) != 1 || string(gotTxs[0].ID) != "tx-id" {
		t.Fatalf("unexpected transactions: %#v", gotTxs)
	}

	active, err := client.AddressIsActive(context.Background(), []byte("pub-key"))
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("expected address to be active")
	}

	if sendErr := client.SendTransaction(context.Background(), *txs[0]); sendErr != nil {
		t.Fatal(sendErr)
	}

	tx, err := client.GetTransactionByID(context.Background(), []byte("tx-id"))
	if err != nil {
		t.Fatal(err)
	}
	if string(tx.ID) != "tx-id" {
		t.Fatalf("tx ID = %q, want %q", tx.ID, "tx-id")
	}
}

func TestClientChainEndpoints(t *testing.T) { //nolint:gocognit // endpoint smoke test is intentionally linear
	encoder := encoding.NewJSONEncoder()

	header := kernel.NewBlockHeader(
		[]byte("v1"),
		123,
		[]byte("merkle-root"),
		7,
		[]byte("previous-block-hash"),
		1,
		42,
	)
	block := kernel.NewBlock(header, []*kernel.Transaction{}, []byte("block-hash"))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1beta/chain/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"height":7,"hash":"626c6f636b2d68617368","timestamp":123}`))
	})
	mux.HandleFunc("/api/v1beta/blocks/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeBlock(*block)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/blocks/626c6f636b2d68617368", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeBlock(*block)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/headers/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeHeader(*header)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/headers/7", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeHeader(*header)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/v1beta/headers", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := encoder.SerializeHeaders([]*kernel.BlockHeader{header})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(data)
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
	if string(gotBlock.Hash) != "block-hash" {
		t.Fatalf("latest block hash = %q, want %q", gotBlock.Hash, "block-hash")
	}

	gotBlock, err = client.GetBlockByHash(context.Background(), []byte("block-hash"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gotBlock.Hash) != "block-hash" {
		t.Fatalf("block hash = %q, want %q", gotBlock.Hash, "block-hash")
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

	gotHeaders, err := client.GetHeaders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(gotHeaders) != 1 || gotHeaders[0].Height != 7 {
		t.Fatalf("unexpected headers: %#v", gotHeaders)
	}
}
