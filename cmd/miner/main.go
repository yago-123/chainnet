package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus/miner"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"chainnet/pkg/wallet"
	"time"

	"github.com/sirupsen/logrus"
)

// temporary constant while the mining difficulty is adjusted
const MiningInterval = 10 * time.Second

func main() {
	var block *kernel.Block
	logger := logrus.New()

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256
	// todo(): add consensusSignatureType

	// create wallet address hasher
	walletSha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {
		logger.Fatalf("Error creating multi-hash configuration: %s", err)
	}

	// create new wallet for storing mining rewards
	w, err := wallet.NewWallet(
		[]byte("1"),
		validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
		crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()),
		walletSha256Ripemd160Hasher,
		hash.GetHasher(consensusHasherType),
	)
	if err != nil {
		logger.Fatalf("Error creating new wallet: %s", err)
	}

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB("bin/miner-storage", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create new mempool
	mempool := miner.NewMemPool()

	// create new observer
	subjectObserver := observer.NewSubjectObserver()

	// create new chain
	chain, err := blockchain.NewBlockchain(
		&config.Config{},
		boltdb,
		hash.GetHasher(consensusHasherType),
		validator.NewHeavyValidator(
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb),
			crypto.NewHashedSignature(
				sign.NewECDSASignature(), hash.NewSHA256(),
			),
			hash.GetHasher(consensusHasherType),
		),
		subjectObserver,
	)
	if err != nil {
		logger.Fatalf("Error creating blockchain: %s", err)
	}

	// create new miner
	mine := miner.NewMiner(w.PublicKey, chain, mempool, hash.SHA256)

	// register chain observers
	subjectObserver.Register(mine)
	subjectObserver.Register(boltdb)
	subjectObserver.Register(mempool)

	for {
		time.Sleep(MiningInterval)

		// start mining block
		block, err = mine.MineBlock()
		if err != nil {
			logger.Errorf("Error mining block: %s", err)
			continue
		}

		if block.IsGenesisBlock() {
			logger.Infof(
				"Genesis block mined successfully: hash %x, number txs: %d, height %d, target %d, nonce %d",
				block.Hash, len(block.Transactions), block.Header.Height, block.Header.Target, block.Header.Nonce,
			)
			continue
		}

		logger.Infof(
			"Block mined successfully: hash %x, previous hash %x, number txs %d, height %d, target %d, nonce %d",
			block.Hash, block.Header.PrevBlockHash, len(block.Transactions), block.Header.Height, block.Header.Target, block.Header.Nonce,
		)
	}
}
