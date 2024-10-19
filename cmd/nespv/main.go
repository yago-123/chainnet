package main

import (
	"github.com/yago-123/chainnet/cmd/nespv/cmd"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	cmd.Execute(logger)

	// initialize network for sending transactions to the miners
	// p2p.NewP2PNode(context.Background(), &config.Config{})

	/*
		// create wallet address hasher
		walletSha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
		if err != nil {
			cfg.Logger.Fatalf("Error creating multi-hash configuration: %s", err)
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
			cfg.Logger.Fatalf("Error creating new wallet: %s", err)
		}
	*/
}
