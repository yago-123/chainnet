package main

import (
	"crypto/sha256"
	"time"

	"github.com/yago-123/chainnet/pkg/monitor"
	"github.com/yago-123/chainnet/pkg/utxoset"

	expl "github.com/yago-123/chainnet/pkg/chain/explorer"

	"github.com/yago-123/chainnet/config"
	blockchain "github.com/yago-123/chainnet/pkg/chain"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/miner"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/storage"

	"github.com/sirupsen/logrus"
)

var cfg *config.Config

var (
	// general consensus hasher (tx, block hashes...)
	consensusHasherType = hash.SHA256

	// general consensus signer (tx)
	consensusSigner = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewHasher(sha256.New()),
	)
)

func main() { //nolint:funlen // main function is expected to have a lot of lines
	var block *kernel.Block

	// execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting chain node with config %v", cfg)

	// create observer controllers
	subjectChain := observer.NewChainSubject()
	subjectNet := observer.NewNetSubject()

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB(cfg.StorageFile, "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	// decorator that wraps storage in order to provide metrics
	meteredBoltdb := storage.NewMeteredStorage(boltdb)

	// create explorer instance
	explorer := expl.NewChainExplorer(meteredBoltdb, hash.GetHasher(consensusHasherType))

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
		meteredBoltdb,
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

	// create new miner
	mine, err := miner.NewMiner(cfg, chain, consensusHasherType, explorer)
	if err != nil {
		cfg.Logger.Fatalf("error initializing miner: %s", err)
	}

	// register network observers
	subjectNet.Register(chain)

	// register chain observers
	subjectChain.Register(mine)
	subjectChain.Register(meteredBoltdb)
	subjectChain.Register(mempool)
	subjectChain.Register(utxoSet)

	network, err := chain.InitNetwork(subjectNet)
	if err != nil {
		cfg.Logger.Fatalf("error initializing network: %s", err)
	}

	// register the block subject to the network
	subjectChain.Register(network)

	// add monitoring via Prometheus
	monitors := []monitor.Monitor{chain, meteredBoltdb, mempool, utxoSet, network, heavyValidator}
	prometheusExporter := monitor.NewPrometheusExporter(cfg, monitors)

	if cfg.Prometheus.Enabled {
		if err = prometheusExporter.Start(); err != nil {
			cfg.Logger.Fatalf("error starting prometheus exporter: %s", err)
		}

		if err == nil {
			cfg.Logger.Infof("exposing Prometheus metrics in http://localhost:%d%s", cfg.Prometheus.Port, cfg.Prometheus.Path)
		}
	}

	for {
		// start mining block
		block, err = mine.MineBlock()
		if err != nil {
			cfg.Logger.Errorf("stopped mining block: %s", err)
			continue
		}

		miningTime := time.Unix(block.Header.Timestamp, 0).Format(time.RFC3339)

		if block.IsGenesisBlock() {
			cfg.Logger.Infof(
				"genesis block mined successfully: hash %x, number txs %d, time %s, height %d, target %d, nonce %d",
				block.Hash, len(block.Transactions), miningTime, block.Header.Height, block.Header.Target, block.Header.Nonce,
			)
			continue
		}

		cfg.Logger.Infof(
			"block mined successfully: hash %x, previous hash %x, number txs %d, time %s, height %d, target %d, nonce %d",
			block.Hash, block.Header.PrevBlockHash, len(block.Transactions), miningTime, block.Header.Height, block.Header.Target, block.Header.Nonce,
		)
	}
}
