package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
	"sort"
	"time"

	"github.com/yago-123/chainnet/pkg/util"
	wallt "github.com/yago-123/chainnet/pkg/wallet/simple_wallet"

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
	MaxNumberConcurrentAccounts = 30
	// MaxNumberWalletsPerAccount is the maximum number of wallets that can be created per account. This limit could be
	// removed, but we don't want to overflow the servers with requests. Each bot will hold 20.000 wallets
	MaxNumberWalletsPerAccount = 5
	FoundationAccountIndex     = 0

	// todo(): make this a flag?
	MetadataPath = "hd_wallet.data"

	MinimumTxBalance = 100000

	TimeoutSendTransaction = 5 * time.Second
	PeriodMetadataBackup   = 1 * time.Minute
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
	logger.SetLevel(logrus.DebugLevel)

	var hdWallet *hd_wallet.Wallet

	// load the wallet "seed"
	privKeyPath := "wallet2.pem"
	privKey, err := util_crypto.ReadECDSAPemToPrivateKeyDerBytes(privKeyPath)
	if err != nil {
		logger.Fatalf("error reading private key: %v", err)
	}

	metadata, err := hd_wallet.LoadMetadata(MetadataPath)
	if err != nil {
		logger.Warnf("error loading metadata: %v", err)
	}

	if metadata == nil {
		// create the hierachical deterministic wallet and sync it
		hdWallet, err = hd_wallet.NewHDWalletWithKeys(
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
		_, err = hdWallet.Sync()
		if err != nil {
			logger.Fatalf("error syncing wallet: %v", err)
		}
	}

	if metadata != nil {
		hdWallet, err = hd_wallet.NewHDWalletWithMetadata(
			cfg,
			1,
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			consensusSigner,
			hash.GetHasher(consensusHasherType),
			encoding.NewProtobufEncoder(),
			privKey,
			metadata,
		)
		if err != nil {
			logger.Fatalf("error initializing HD wallet with metadata: %v", err)
		}
	}

	numAccounts := hdWallet.GetNumAccounts()
	logger.Infof("HD wallet initialized, contains %d accounts", numAccounts)

	// save the metadata periodically
	go SaveMetadataPeriodically(hdWallet)

	logger.Infof("HD wallet synced with %d accounts", numAccounts)

	// if there are no active accounts, ask for funds and exit the program
	if numAccounts == 0 {
		if errAskFunds := AskForFunds(hdWallet); errAskFunds != nil {
			logger.Fatalf("error asking for funds: %v", errAskFunds)
		}

		return
	}

	// create remaining accounts if needed so that we can operate them in parallel without problems
	numAccounts = hdWallet.GetNumAccounts()
	if numAccounts < MaxNumberConcurrentAccounts {
		logger.Infof("creating remaining %d accounts...", MaxNumberConcurrentAccounts-numAccounts)
		for i := numAccounts; i < MaxNumberConcurrentAccounts; i++ {
			_, errAccount := hdWallet.GetNewAccount()
			if errAccount != nil {
				logger.Fatalf("error creating account: %v", errAccount)
			}
		}
	}

	// distribute funds among accounts regardless of the number of accounts. This is done so that we can refill
	// the bots by transfering funds to the foundation account and restarting the bot
	if errDistrFund := DistributeFundsAmongAccounts(hdWallet); errDistrFund != nil {
		logger.Warnf("error distributing funds from foundation account: %v", errDistrFund)
	}

	// distribute funds between wallets for each account (isolated)
	for i := 0; i < MaxNumberConcurrentAccounts; i++ {
		// skip the foundation account
		if i == FoundationAccountIndex {
			continue
		}

		// retrieve account and start the generation of transactions among the wallets contained
		account, errAcc := hdWallet.GetAccount(uint(i))
		if errAcc != nil {
			logger.Fatalf("error getting account: %v", errAcc)
		}
		go DistributeFundsBetweenWallets(account)
	}

	select {}
}

// AskForFunds asks for funds to the user by displaying the P2PK address of the first wallet in the HD wallet
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

// DistributeFundsAmongAccounts distributes the funds from the foundation account to the rest of the accounts in the
// HD wallet. This is done so that accounts can start operating in an isolated way without having to rely on external
// funds
func DistributeFundsAmongAccounts(hdWallet *hd_wallet.Wallet) error {
	addresses := [][]byte{}
	targetAmounts := []uint{}

	totalBalance, err := hdWallet.GetBalance()
	if err != nil {
		return fmt.Errorf("error getting wallet balance: %w", err)
	}

	logger.Infof("HD wallet contains %.5f coins", float64(totalBalance)/float64(kernel.ChainnetCoinAmount))

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
	distributeFundsAmount := (foundationAccountBalance) / (MaxNumberConcurrentAccounts + 1)
	for i := range MaxNumberConcurrentAccounts {
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

	logger.Infof("funds distributed to %d accounts: %s", MaxNumberConcurrentAccounts, tx.String())

	return nil
}

func DistributeFundsBetweenWallets(acc *hd_wallet.Account) {
	var tx *kernel.Transaction
	var amounts []uint
	var addresses [][]byte

	logrus.Infof("starting funds distribution for account %d", acc.GetAccountID())
	for {
		// sleep randomized so that the nodes are not overflowed
		time.Sleep(time.Duration(rand.UintN(200)+20) * time.Second)

		// get the UTXOs of the account
		accUTXOs, err := acc.GetAccountUTXOs()
		if err != nil {
			logger.Warnf("error getting account UTXOs for account %d: %v", acc.GetAccountID(), err)
		}

		if len(accUTXOs) == 0 {
			logger.Warnf("no UTXOs found for account %d, skipping execution", acc.GetAccountID())
			continue
		}

		// if there are less than 10 UTXOs, then generate multi outputs transaction (>15 outputs)
		if len(accUTXOs) < 10 {
			for _, utxo := range accUTXOs {
				addresses = GetRandomAccountAddresses(1, MaxNumberWalletsPerAccount, acc)
				amounts = GetRandomAmounts(utxo.GetAmount(), uint(len(addresses))+1) // add 1 for the tx fees
				tx, err = acc.GenerateNewTransaction(script.P2PKH, addresses, amounts[:len(amounts)-1], amounts[len(amounts)-1], []*kernel.UTXO{utxo})
				if err != nil {
					logger.Warnf("error generating transaction: %v", err)
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), TimeoutSendTransaction)
				defer cancel()

				if errSend := acc.SendTransaction(ctx, tx); errSend != nil {
					logger.Warnf("error sending transaction: %v", errSend)
					continue
				}

				logger.Debugf("account %d distributed %f coins to %d addresses: %s",
					acc.GetAccountID(),
					kernel.ConvertFromChannoshisToCoins(utxo.GetAmount()),
					len(addresses),
					tx.String())

				// sleep after each transaction is sent to avoid overflow
				time.Sleep(time.Duration(rand.UintN(200)+10) * time.Second)
			}

			continue
		}

		// if there are more than 10 UTXOs, then split the UTXOs array in 5 and do periodic transactions
		splitedUTXOs := util.SplitArray(accUTXOs, 5)
		for _, utxos := range splitedUTXOs {
			totalBalanceTx := util.GetBalanceUTXOs(utxos)
			if totalBalanceTx < MinimumTxBalance {
				// if the balance is too small, send the transaction with a single output
				addresses = GetRandomAccountAddresses(0, 1, acc)
				amounts = GetRandomAmounts(totalBalanceTx, uint(len(addresses))+1)
			}

			if totalBalanceTx > MinimumTxBalance {
				addresses = GetRandomAccountAddresses(1, MaxNumberWalletsPerAccount, acc)
				amounts = GetRandomAmounts(totalBalanceTx, uint(len(addresses))+1)
			}

			tx, err = acc.GenerateNewTransaction(script.P2PKH, addresses, amounts[:len(amounts)-1], amounts[len(amounts)-1], utxos)
			if err != nil {
				logger.Warnf("error generating transaction: %v", err)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), TimeoutSendTransaction)
			defer cancel()

			if errSend := acc.SendTransaction(ctx, tx); errSend != nil {
				logger.Warnf("error sending transaction: %v", errSend)
				continue
			}

			logger.Debugf("account %d distributed %f coins to %d addresses: %s",
				acc.GetAccountID(),
				kernel.ConvertFromChannoshisToCoins(util.GetBalanceUTXOs(utxos)),
				len(addresses),
				tx.String())

			time.Sleep(time.Duration(rand.UintN(60)+30) * time.Second)
		}
	}
}

func GetRandomAmounts(totalBalance, numAddresses uint) []uint {
	if numAddresses == 0 || totalBalance == 0 {
		return []uint{}
	}

	// generate N-1 random points in the range [0, totalBalance]
	randomPoints := make([]uint, numAddresses-1)
	for i := range randomPoints {
		randomPoints[i] = uint(rand.UintN(totalBalance + 1))
	}

	// sort the random points to create ranges
	sort.Slice(randomPoints, func(i, j int) bool { return randomPoints[i] < randomPoints[j] })

	// calculate balances as differences between sorted random points
	balances := make([]uint, numAddresses)
	prev := uint(0)
	for i, point := range randomPoints {
		balances[i] = point - prev
		prev = point
	}
	balances[numAddresses-1] = totalBalance - prev

	return balances
}

func GetRandomAccountAddresses(min, max uint, account *hd_wallet.Account) [][]byte {
	var err error
	var addresses [][]byte

	numAddresses := rand.UintN(max-min) + min

	for i := uint(0); i < numAddresses; i++ {
		var wallet *wallt.Wallet

		// if the limit have been reached, pick an existing wallet
		if account.GetExternalWalletIndex() >= MaxNumberWalletsPerAccount {
			wallet, err = account.GetExternalWallet(rand.UintN(MaxNumberWalletsPerAccount))
			if err != nil {
				logger.Warnf("error getting external wallet: %v", err)
				continue
			}
		}

		// if not all wallets have been created, create a new one
		if account.GetExternalWalletIndex() < MaxNumberWalletsPerAccount {
			wallet, err = account.GetNewExternalWallet()
			if err != nil {
				logger.Warnf("error getting new wallet: %v", err)
				continue
			}
		}

		address, errAddress := wallet.GetP2PKHAddress()
		if errAddress != nil {
			logger.Warnf("error getting P2PKH address: %v", errAddress)
			continue
		}

		addresses = append(addresses, address)
	}

	return addresses
}

func SaveMetadataPeriodically(hdWallet *hd_wallet.Wallet) {
	for {
		time.Sleep(PeriodMetadataBackup)
		hd_wallet.SaveMetadata("hd_wallet.data", hdWallet.GetMetadata())
	}
}
