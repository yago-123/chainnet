package main

import (
	"crypto/sha256"

	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
	"github.com/yago-123/chainnet/pkg/wallet/hd"
)

var (
	// general consensus hasher (tx, block hashes...)
	consensusHasherType = hash.SHA256

	// general consensus signer (tx)
	consensusSigner = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewHasher(sha256.New()),
	)
)

var logger = logrus.New()
var cfg = config.NewConfig()

func main() {
	privKeyPath := "wallet.pem"
	var utxos []*kernel.UTXO

	privKey, err := util_crypto.ReadECDSAPemToPrivateKeyDerBytes(privKeyPath)
	if err != nil {
		logger.Fatalf("error reading private key: %v", err)
	}

	hdWallet, err := hd.NewHDWalletWithKeys(
		cfg,
		1,
		validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
		consensusSigner,
		hash.GetHasher(consensusHasherType),
		encoding.NewProtobufEncoder(),
		privKey,
	)
	if err != nil {
		logger.Fatalf("error initializing HD wallet: %v", err)
	}

	numAccounts, err := hdWallet.Sync()
	if err != nil {
		logger.Fatalf("error syncing wallet: %v", err)
	}

	for i := 0; i < int(numAccounts); i++ { //nolint:gosec,intrange // possibility of integer overflow is OK here
		// create a new account
		hda, errHda := hdWallet.GetAccount(uint(i))
		if errHda != nil {
			logger.Fatalf("error getting new account: %v", errHda)
		}

		// create a new wallet
		wallet, errWallet := hda.GetNewWallet()
		if errWallet != nil {
			logger.Fatalf("error generating wallet for account %d new wallet: %v", i, errWallet)
		}

		// initialize the network for the wallet
		_, err = wallet.InitNetwork()
		if err != nil {
			logger.Fatalf("error setting up wallet network: %v", err)
		}

		// get the wallet UTXOS
		utxos, err = wallet.GetWalletUTXOS()
		if err != nil {
			logger.Fatalf("error getting wallet UTXOS: %v", err)
		}

		rawPubKey, errRaw := util_crypto.DecodeDERBytesToRawPublicKey(wallet.GetP2PKAddress())
		if errRaw != nil {
			logger.Fatalf("error getting raw public key: %v", err)
		}
		logger.Infof("account number: %d, account first wallet address: %x, number of utxos: %d", hda.GetAccountID(), rawPubKey, len(utxos))
	}
}
