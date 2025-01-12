package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
	"github.com/yago-123/chainnet/pkg/wallet/hd_wallet"
)

const (
	ConcurrentAccounts     = 4
	FoundationAccountIndex = 0

	MinimumTxBalance = 100

	SleepTimeBetweenRecalculations = 20 * time.Minute
	TimeBetweenMetadataBackup      = 1 * time.Minute
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
	privKeyPath := "wallet2.pem"
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

	go SaveMetadataPeriodically(hdWallet)

	logger.Infof("syncing HD wallet...")
	numAccounts, err := hdWallet.Sync()
	if err != nil {
		logger.Fatalf("error syncing wallet: %v", err)
	}

	logger.Infof("HD wallet synced with %d accounts", numAccounts)

	if numAccounts == 0 {
		if errAskFunds := AskForFunds(hdWallet); errAskFunds != nil {
			logger.Fatalf("error asking for funds: %v", errAskFunds)
		}
	}

	if numAccounts == 1 {
		if errDistrFund := DistributeFundsAmongAccounts(hdWallet); errDistrFund != nil {
			logger.Fatalf("error distributing funds: %v", errDistrFund)
		}
	}

	if numAccounts > 1 {
		// keep redistributing funds
		// if account exists but wallets is 1
		CreateMultipleWalletsInsideAccount()
	}

	// if numAccounts > 1 && numWallets > 1 {
	// 	CreateTransactionsInsideAccount()
	// }
}

func AskForFunds(hdWallet *hd_wallet.Wallet) error {
	acc, errAcc := hdWallet.GetNewAccount()
	if errAcc != nil {
		return fmt.Errorf("error getting account: %w", errAcc)
	}

	wallet, errAcc := acc.GetNewExternalWallet()
	if errAcc != nil {
		return fmt.Errorf("error getting wallet: %w", errAcc)
	}

	logger.Warnf("HD wallet is empty, fund %s with a P2PK and execute this again", base58.Encode(wallet.GetP2PKAddress()))

	return nil
}

func DistributeFundsAmongAccounts(hdWallet *hd_wallet.Wallet) error {
	addresses := [][]byte{}
	targetAmounts := []uint{}

	totalBalance, err := hdWallet.GetBalance()
	if err != nil {
		return fmt.Errorf("error getting wallet balance: %w", err)
	}

	logger.Infof("HD wallet contains %.5f coins", float64(totalBalance)/float64(kernel.ChainnetCoinAmount))

	numAccounts := hdWallet.GetNumAccounts()

	// create remaining accounts so that we can operate them in parallel without problems
	if numAccounts < ConcurrentAccounts {
		logger.Infof("creating remaining %d accounts...", ConcurrentAccounts-numAccounts)
		for i := numAccounts; i < ConcurrentAccounts; i++ {
			_, errAccount := hdWallet.GetNewAccount()
			if errAccount != nil {
				return fmt.Errorf("error creating account: %w", errAccount)
			}
		}
	}

	// check the balance of the foundation account
	foundationAccount, err := hdWallet.GetAccount(FoundationAccountIndex)
	if err != nil {
		return fmt.Errorf("error getting foundation account: %w", err)
	}

	foundationAccountBalance, err := foundationAccount.GetBalance()
	if err != nil {
		return fmt.Errorf("error getting foundation account balance: %w", err)
	}

	logger.Infof("foundation account contains %.5f coins", kernel.ConvertFromChannoshisToCoins(foundationAccountBalance))

	// generate outputs for multiple addresses
	distributeFundsAmount := (foundationAccountBalance) / (ConcurrentAccounts + 1)
	for i := range ConcurrentAccounts {
		targetAmounts = append(targetAmounts, distributeFundsAmount)

		account, errAccount := hdWallet.GetAccount(uint(i))
		if errAccount != nil {
			return fmt.Errorf("error getting account: %w", errAccount)
		}

		wallet, errWallet := account.GetNewExternalWallet()
		if errWallet != nil {
			return fmt.Errorf("error getting wallet: %w", errWallet)
		}

		addresses = append(addresses, wallet.GetP2PKAddress())
	}

	foundationAccountUTXOs, err := foundationAccount.GetAccountUTXOs()
	if err != nil {
		return fmt.Errorf("error getting foundation account UTXOs: %w", err)
	}

	// create the foundation fund transaction
	tx, err := foundationAccount.GenerateNewTransaction(
		script.P2PK,
		addresses,
		targetAmounts,
		distributeFundsAmount,
		foundationAccountUTXOs,
	)
	if err != nil {
		return fmt.Errorf("error generating transaction: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.P2P.ConnTimeout)
	defer cancel()

	if errSend := foundationAccount.SendTransaction(ctx, tx); errSend != nil {
		return fmt.Errorf("error sending transaction: %w", errSend)
	}

	logger.Infof("funds distributed to %d accounts: %s", ConcurrentAccounts, tx.String())

	return nil
}

func CreateMultipleWalletsInsideAccount() {

}

func CreateTransactionsInsideAccount() {

}

func SaveMetadataPeriodically(hdWallet *hd_wallet.Wallet) {
	for {
		time.Sleep(TimeBetweenMetadataBackup)
		hd_wallet.SaveMetadata("hd_wallet.data", hdWallet.GetMetadata())
	}
}
