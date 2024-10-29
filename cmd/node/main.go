package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	blockchain "github.com/yago-123/chainnet/pkg/chain"
	expl "github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/monitor"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/storage"
	"github.com/yago-123/chainnet/pkg/utxoset"
)

var cfg *config.Config

var (
	// general consensus hasher (tx, block hashes...)
	consensusHasherType = hash.SHA256

	// general consensus signer (tx)
	consensusSigner = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewSHA256(),
	)
)

func main() {
	var err error

	// execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting chain node with config %v", cfg)

	// create new observer
	netSubject := observer.NewNetSubject()
	subjectChain := observer.NewChainSubject()

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB(cfg.StorageFile, "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create explorer instance
	explorer := expl.NewExplorer(boltdb, hash.GetHasher(consensusHasherType))

	// create mempool instance
	mempool := mempool.NewMemPool(cfg.Chain.MaxTxsMempool)

	// create utxo set instance
	utxoSet := utxoset.NewUTXOSet(cfg)

	// create heavy validator
	heavyValidator := validator.NewHeavyValidator(
		cfg,
		validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
		explorer,
		consensusSigner,
		hash.GetHasher(consensusHasherType),
	)

	// define encoder type
	encoder := encoding.NewProtobufEncoder()

	// create new chain
	chain, err := blockchain.NewBlockchain(
		cfg,
		boltdb,
		mempool,
		utxoSet,
		hash.GetHasher(consensusHasherType),
		heavyValidator,
		subjectChain,
		encoder,
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating blockchain: %s", err)
	}

	// register network observers
	netSubject.Register(chain)

	// register chain observers
	subjectChain.Register(boltdb)
	subjectChain.Register(mempool)
	subjectChain.Register(utxoSet)

	network, err := chain.InitNetwork(netSubject)
	if err != nil {
		cfg.Logger.Fatalf("Error initializing network: %s", err)
	}

	// register the block subject to the network
	subjectChain.Register(network)

	// add monitoring via Prometheus
	monitors := []monitor.Monitor{chain, boltdb, mempool, utxoSet}
	prometheusExporter := monitor.NewPrometheusExporter(cfg, monitors)

	if cfg.Prometheus.Enabled {
		if err = prometheusExporter.Start(); err != nil {
			cfg.Logger.Fatalf("Error starting prometheus exporter: %s", err)
		}

		if err == nil {
			cfg.Logger.Infof("Exposing Prometheus metrics in localhost:%d%s", cfg.Prometheus.Port, cfg.Prometheus.Path)
		}
	}

	select {}
}
