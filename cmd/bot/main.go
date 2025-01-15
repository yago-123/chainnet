package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
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
	MaxNumberConcurrentAccounts = 5
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
	cfg.Wallet.ServerAddress = "127.0.0.1"
	cfg.Wallet.ServerPort = 8080

	logger.SetLevel(logrus.DebugLevel)

	var hdWallet *hd_wallet.Wallet

	// load the wallet "seed"
	privKeyPath := "wallet-test.pem"
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
	// go SaveMetadataPeriodically(hdWallet)

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

	/*
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

	*/

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

	var wallet *wallt.Wallet

	totalBalance, err := hdWallet.GetBalance()
	if err != nil {
		return fmt.Errorf("error getting wallet balance: %w", err)
	}

	logger.Infof("HD wallet contains %.5f coins", kernel.ConvertFromChannoshisToCoins(totalBalance))

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

	if foundationAccountBalance < MinimumTxBalance {
		logger.Warnf("foundation account balance is below the minimum transaction balance, skipping distribution")
		return nil
	}

	// generate one output for each account by distributing the funds equally
	distributeFundsAmount := foundationAccountBalance / MaxNumberConcurrentAccounts
	for i := range MaxNumberConcurrentAccounts {
		targetAmounts = append(targetAmounts, distributeFundsAmount)

		// retrieve the account and choose the wallet to send the funds to
		account, errAccount := hdWallet.GetAccount(uint(i))
		if errAccount != nil {
			return fmt.Errorf("error getting account: %w", errAccount)
		}

		if account.GetExternalWalletIndex() >= MaxNumberWalletsPerAccount {
			wallet, err = account.GetExternalWallet(rand.UintN(MaxNumberWalletsPerAccount))
			if err != nil {
				return fmt.Errorf("error getting external wallet: %w", err)
			}
		}

		if account.GetExternalWalletIndex() < MaxNumberWalletsPerAccount {
			wallet, err = account.GetNewExternalWallet()
			if err != nil {
				return fmt.Errorf("error getting new wallet: %w", err)
			}
		}

		// get the P2PKH address of the wallet
		address, errWallet := wallet.GetP2PKHAddress()
		if errWallet != nil {
			return fmt.Errorf("error getting P2PKH address: %w", errWallet)
		}

		addresses = append(addresses, address)
	}

	// retrieve the UTXOs and create the whole transaction
	foundationAccountUTXOs, err := foundationAccount.GetAccountUTXOs()
	if err != nil {
		return fmt.Errorf("error getting foundation account UTXOs: %w", err)
	}

	if err = createAndSendTransaction(foundationAccount, addresses, targetAmounts, 0, foundationAccountUTXOs); err != nil {
		return fmt.Errorf("error creating and sending transaction: %w", err)
	}

	return nil
}

/*
// DistributeFundsBetweenWallets distributes the funds between the wallets of an account. This is done so that the
// account can operate in an isolated way without having to rely on external funds (until the tx fees waste all the
// funds)
func DistributeFundsBetweenWallets(acc *hd_wallet.Account) {
	logrus.Infof("Starting funds distribution for account %d", acc.GetAccountID())

	for {
		time.Sleep(randomizedSleep(20, 220))

		accUTXOs, err := acc.GetAccountUTXOs()
		if err != nil {
			logger.Warnf("Error getting UTXOs for account %d: %v", acc.GetAccountID(), err)
			continue
		}

		// if the account has no UTXOs, skip the execution
		if len(accUTXOs) == 0 {
			logger.Warnf("No UTXOs found for account %d, skipping execution", acc.GetAccountID())
			continue
		}

		// if the account has a small number of UTXOs, distribute them to a small number of wallets
		if len(accUTXOs) < 10 {
			// distributeSmallUTXOs(acc, accUTXOs)
			// continue
		}

		// if the account has a large number of UTXOs, split them into smaller groups
		distributeLargeUTXOs(acc, accUTXOs)
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

func GetRandomAccountAddress(account *hd_wallet.Account) ([]byte, error) {
	var err error
	var wallet *wallt.Wallet

	// if the limit have been reached, pick an existing wallet
	if account.GetExternalWalletIndex() >= MaxNumberWalletsPerAccount {
		wallet, err = account.GetExternalWallet(rand.UintN(MaxNumberWalletsPerAccount))
		if err != nil {
			return []byte{}, fmt.Errorf("error getting external wallet: %v", err)
		}
	}

	// if not all wallets have been created, create a new one
	if account.GetExternalWalletIndex() < MaxNumberWalletsPerAccount {
		wallet, err = account.GetNewExternalWallet()
		if err != nil {
			return []byte{}, fmt.Errorf("error getting new wallet: %v", err)
		}
	}

	address, errAddress := wallet.GetP2PKHAddress()
	if errAddress != nil {
		return []byte{}, fmt.Errorf("error getting P2PKH address: %v", errAddress)
	}

	return address, nil
}

func GetRandomAccountAddresses(min, max uint, account *hd_wallet.Account) [][]byte {
	var err error
	var addresses [][]byte

	numAddresses := rand.UintN(max-min) + min

	for i := uint(0); i < numAddresses; i++ {
		address, errAddress := GetRandomAccountAddress(account)
		if errAddress != nil {
			logger.Warnf("error getting random account address: %v", err)
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

func distributeSmallUTXOs(acc *hd_wallet.Account, accUTXOs []*kernel.UTXO) {
	for _, utxo := range accUTXOs {
		addresses := GetRandomAccountAddresses(1, MaxNumberWalletsPerAccount-1, acc)
		amounts := GetRandomAmounts(utxo.Amount(), uint(len(addresses)))
		if err := createAndSendTransaction(acc, addresses, amounts, 0, []*kernel.UTXO{utxo}); err != nil {
			logger.Warnf("error creating and sending transaction from distributeSmallUTXOs: %v", err)
		}

		time.Sleep(randomizedSleep(10, 210))
	}
}

func distributeLargeUTXOs(acc *hd_wallet.Account, accUTXOs []*kernel.UTXO) {
	splitUTXOs := util.SplitArray(accUTXOs, 5)

	for _, utxos := range splitUTXOs {
		totalBalance := util.GetBalanceUTXOs(utxos)

		if totalBalance < MinimumTxBalance {
			address, err := GetRandomAccountAddress(acc)
			if err != nil {
				logger.Warnf("error getting random account address: %v", err)
				continue
			}

			logger.Infof("created address: %s", base58.Encode(address))
			if errSend := createAndSendTransaction(acc, [][]byte{address}, []uint{totalBalance}, 0, utxos); errSend != nil {
				logger.Warnf("error creating and sending transaction from distributeLargeUTXOs 1: %v", errSend)
			}
		}

		if totalBalance >= MinimumTxBalance {
			addresses := GetRandomAccountAddresses(1, MaxNumberWalletsPerAccount-1, acc)
			for _, address := range addresses {
				logger.Infof("created address: %s", base58.Encode(address))
			}
			amounts := GetRandomAmounts(totalBalance, uint(len(addresses)))
			if errSend := createAndSendTransaction(acc, addresses, amounts, 0, utxos); errSend != nil {
				logger.Warnf("error creating and sending transaction from distributeLargeUTXOs 2: %v", errSend)
			}
		}

		time.Sleep(randomizedSleep(30, 90))
	}
}

func randomizedSleep(min, max uint) time.Duration {
	return time.Duration(rand.UintN(max-min)+min) * time.Second
}
*/

func createAndSendTransaction(acc *hd_wallet.Account, addresses [][]byte, amounts []uint, txFee uint, utxos []*kernel.UTXO) error {
	var err error
	var wallet *wallt.Wallet

	if acc.GetInternalWalletIndex() == 0 {
		wallet, err = acc.GetNewInternalWallet()
	} else {
		wallet, err = acc.GetInternalWallet(0)
	}

	if err != nil {
		return fmt.Errorf("error getting internal wallet: %v", err)
	}

	tx, err := acc.GenerateNewTransaction(
		script.P2PKH,
		addresses,
		amounts,
		txFee,
		wallet.PublicKey(),
		1,
		utxos,
	)
	if err != nil {
		return fmt.Errorf("error generating transaction (create and send): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutSendTransaction)
	defer cancel()

	if errSend := acc.SendTransaction(ctx, tx); errSend != nil {
		return fmt.Errorf("error sending transaction: %v", errSend)
	}

	logger.Debugf("account %d distributed %f coins to %d addresses: %s",
		acc.GetAccountID(),
		kernel.ConvertFromChannoshisToCoins(util.GetBalanceUTXOs(utxos)),
		len(addresses),
		tx.String(),
	)

	return nil
}