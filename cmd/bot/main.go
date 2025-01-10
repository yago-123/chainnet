package main

import (
	"crypto/sha256"
	"github.com/btcsuite/btcutil/base58"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
	"github.com/yago-123/chainnet/pkg/wallet/hd_wallet"
)

const (
	ConcurrentAccounts     = 100
	FoundationAccountIndex = 0

	MinimumTxBalance = 100

	SleepTimeBetweenRecalculations = 20 * time.Minute
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
	cfg.Logger.SetLevel(logrus.DebugLevel)

	// load the "seed"
	privKeyPath := "wallet.pem"
	privKey, err := util_crypto.ReadECDSAPemToPrivateKeyDerBytes(privKeyPath)
	if err != nil {
		logger.Fatalf("error reading private key: %v", err)
	}

	// create the hierachical deterministic wallet and sync it
	hdWallet, err := hd_wallet.NewHDWalletWithKeys(
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

	logger.Infof("syncing HD wallet...")
	numAccounts, err := hdWallet.Sync()
	if err != nil {
		logger.Fatalf("error syncing wallet: %v", err)
	}

	logger.Infof("HD wallet synced with %d accounts", numAccounts)

	totalBalance, err := hdWallet.GetBalance()
	if err != nil {
		logger.Fatalf("error getting wallet balance: %v", err)
	}

	if totalBalance == 0 {
		acc, errAcc := hdWallet.GetAccount(FoundationAccountIndex)
		if errAcc != nil {
			logger.Fatalf("error getting foundation account: %v", errAcc)
		}

		wallet, errAcc := acc.GetNewExternalWallet()
		if errAcc != nil {
			logger.Fatalf("error getting foundation wallet: %v", errAcc)
		}

		logger.Fatalf("HD wallet is empty, fund %s with a P2PK and execute this again", base58.Encode(wallet.GetP2PKAddress()))
	}

	logger.Infof("HD wallet contains %.5f coins", float64(totalBalance)/float64(kernel.ChainnetCoinAmount))

	// create remaining accounts so that we can operate them in parallel without problems
	if numAccounts < ConcurrentAccounts {
		logger.Infof("creating remaining %d accounts...", ConcurrentAccounts-numAccounts)
		for i := numAccounts; i < ConcurrentAccounts; i++ {
			_, errAccount := hdWallet.GetNewAccount()
			if errAccount != nil {
				logger.Fatalf("error creating account: %v", errAccount)
			}
		}
	}

	// check the balance of the foundation account
}
