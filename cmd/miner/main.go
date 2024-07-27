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

	"github.com/sirupsen/logrus"
)

func main() {
	var block *kernel.Block
	logger := logrus.New()

	// create wallet address hasher
	sha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {
		logger.Fatalf("Error creating multi-hash configuration: %s", err)
	}

	// create new wallet for storing mining rewards
	w, err := wallet.NewWallet(
		[]byte("1"),
		validator.NewLightValidator(hash.NewSHA256()),
		crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()),
		sha256Ripemd160Hasher,
	)
	if err != nil {
		logger.Fatalf("Error creating new wallet: %s", err)
	}

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB("boltdb-file", "block-bucket", "header-bucket", encoding.NewGobEncoder())
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
		validator.NewHeavyValidator(validator.NewLightValidator(sha256Ripemd160Hasher), explorer.NewExplorer(boltdb), crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()), sha256Ripemd160Hasher),
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
		// start mining block
		block, err = mine.MineBlock()
		if err != nil {
			logger.Errorf("Error mining block: %s", err)
			continue
		}

		logger.Infof("Mined block: %s", block.Hash)
	}
}
