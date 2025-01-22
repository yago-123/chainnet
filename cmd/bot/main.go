package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
	"sort"
	"time"

	hd_wallet "github.com/yago-123/chainnet/pkg/wallet/hd_wallet"

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

	// limits for startup of fund distribution, prevents all bots asking for funds at the same time
	MinTimeStartupFundDistribution = 1 * time.Second
	MaxTimeStartupFundDistribution = 500 * time.Second

	// limits for time between transactions, apply at account level
	MinTimeBetweenTransactions = 60 * time.Second
	MaxTimeBetweenTransactions = 200 * time.Second

	TimeoutSendTransaction = 10 * time.Second
	PeriodMetadataBackup   = 1 * time.Minute

	// MaxInputGroupsForCreatingTx is the maximum number of input groups that can be used for creating a transaction
	MaxInputGroupsForCreatingTx = 4
	// MaxOutputGroupsForCreatingTx is the maximum number of output groups that can be used for creating a transaction
	// must always be smaller than MaxNumberWalletsPerAccount
	MaxOutputGroupsForCreatingTx = 4
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

func main() { //nolint:funlen,gocognit,nolintlint // this is a main function, it's OK to be long here
	cfg.Logger.SetLevel(logrus.DebugLevel)
	logger.SetLevel(logrus.DebugLevel)

	var hdWallet *hd_wallet.Wallet

	// load the wallet "seed"
	privKeyPath := "wallet.pem"
	privKey, err := util_crypto.ReadECDSAPemToPrivateKeyDerBytes(privKeyPath)
	if err != nil {
		logger.Fatalf("error reading private key: %v", err)
	}

	metadata, err := hd_wallet.LoadMetadata(MetadataPath)
	if err != nil {
		logger.Warnf("error loading metadata: %v", err)
	}

	if MaxOutputGroupsForCreatingTx > MaxNumberWalletsPerAccount {
		logger.Fatalf("MaxOutputGroupsForCreatingTx must be smaller or equal than MaxNumberWalletsPerAccount")
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
	for i := range MaxNumberConcurrentAccounts {
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
	var wallet *wallt.Wallet
	acc, err := hdWallet.GetNewAccount()
	if err != nil {
		return fmt.Errorf("error getting account: %w", err)
	}

	if acc.GetExternalWalletIndex() > 0 {
		wallet, err = acc.GetExternalWallet(0)
		if err != nil {
			return fmt.Errorf("error getting external wallet: %w", err)
		}
	}
	if acc.GetExternalWalletIndex() == 0 {
		wallet, err = acc.GetNewExternalWallet()
		if err != nil {
			return fmt.Errorf("error getting wallet: %w", err)
		}
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

	// check the balance of the foundation account
	foundationAccount, err := hdWallet.GetAccount(FoundationAccountIndex)
	if err != nil {
		return fmt.Errorf("error getting foundation account: %w", err)
	}

	foundationAccountBalance, err := foundationAccount.GetBalance()
	if err != nil {
		return fmt.Errorf("error getting foundation account balance: %w", err)
	}

	logger.Infof("foundation account contains approx. %.5f coins", kernel.ConvertFromChannoshisToCoins(foundationAccountBalance))

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

		// if the limit have been reached, pick an existing random wallet
		if account.GetExternalWalletIndex() >= MaxNumberWalletsPerAccount {
			wallet, err = account.GetExternalWallet(rand.UintN(MaxNumberWalletsPerAccount)) //nolint:gosec // no need for secure random here
			if err != nil {
				return fmt.Errorf("error getting external wallet: %w", err)
			}
		}

		// if the limit have not been reached, create a new wallet
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

// DistributeFundsBetweenWallets distributes the funds between the wallets of an account. This is done so that the
// account can operate in an isolated way without having to rely on external funds (until the tx fees waste all the
// funds)
func DistributeFundsBetweenWallets(acc *hd_wallet.Account) { //nolint:funlen,gocognit,nolintlint // this is a core func for bot, it's OK to be long here
	logrus.Infof("starting funds distribution for account %d", acc.GetAccountID())

	// sleep for a random amount of time before starting the distribution so that we avoid all accounts asking
	// for UTXOs at the same time
	randomizedSleep(MinTimeStartupFundDistribution, MaxTimeStartupFundDistribution)

	for {
		accUTXOs, err := acc.GetAccountUTXOs()
		if err != nil {
			logger.Warnf("error getting UTXOs for account %d: %v", acc.GetAccountID(), err)
			randomizedSleep(MinTimeBetweenTransactions, MaxTimeBetweenTransactions)
			continue
		}

		// if the account has no UTXOs, skip the execution
		if len(accUTXOs) == 0 {
			logger.Warnf("no UTXOs found for account %d, skipping execution", acc.GetAccountID())
			randomizedSleep(MinTimeBetweenTransactions, MaxTimeBetweenTransactions)

			continue
		}

		for _, utxos := range splitArrayRandomized(accUTXOs, MaxInputGroupsForCreatingTx) {
			addresses := [][]byte{}

			// if the utxo is small than the minimum transaction balance, send it to a single address
			if util.GetBalanceUTXOs(utxos) <= MinimumTxBalance {
				addr, errTmp := getRandomAccountAddress(acc)
				if errTmp != nil {
					logger.Warnf("error getting random account address: %v", errTmp)
					continue
				}

				addresses = append(addresses, addr)
			}

			// otherwise, split the UTXO in a number of random addresses and amounts
			if util.GetBalanceUTXOs(utxos) > MinimumTxBalance {
				addresses, err = getRandomAccountAddresses(1, MaxOutputGroupsForCreatingTx, acc)
				if err != nil {
					logger.Warnf("error getting random account addresses: %v", err)
					continue
				}
			}

			// create and send the transaction
			amounts := getRandomAmounts(util.GetBalanceUTXOs(utxos), uint(len(addresses)+1)) // add one for the tx fee
			if errTx := createAndSendTransaction(acc, addresses, amounts[:len(amounts)-1], amounts[len(amounts)-1], utxos); errTx != nil {
				logger.Warnf("error creating and sending transaction: %v", errTx)
			}

			randomizedSleep(MinTimeBetweenTransactions, MaxTimeBetweenTransactions)
		}
	}
}

// SaveMetadataPeriodically saves the metadata of the HD wallet periodically
func SaveMetadataPeriodically(hdWallet *hd_wallet.Wallet) {
	for {
		time.Sleep(PeriodMetadataBackup)
		if err := hd_wallet.SaveMetadata(MetadataPath, hdWallet.GetMetadata()); err != nil {
			logger.Warnf("error saving metadata: %v", err)
		}
	}
}

// getRandomAmounts generates random amounts to be distributed among a number of addresses. The total balance is split
func getRandomAmounts(totalBalance, numAddresses uint) []uint {
	if numAddresses == 0 || totalBalance == 0 {
		return []uint{}
	}

	// generate N-1 random points in the range [0, totalBalance]
	randomPoints := make([]uint, numAddresses-1)
	for i := range randomPoints {
		randomPoints[i] = rand.UintN(totalBalance + 1) //nolint:gosec // no need for secure random here
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

func randomizedSleep(minDuration, maxDuration time.Duration) {
	// do nothing if the range is invalid
	if maxDuration <= minDuration {
		return
	}

	durationRange := maxDuration - minDuration
	// generate a random duration within the range
	randomDuration := minDuration + rand.N(durationRange) //nolint:gosec // no need for secure random here
	time.Sleep(randomDuration)
}

func createAndSendTransaction(acc *hd_wallet.Account, addresses [][]byte, amounts []uint, txFee uint, utxos []*kernel.UTXO) error {
	var err error
	var wallet *wallt.Wallet

	// retrieve change address by retrieving a internal wallet from the account
	if acc.GetInternalWalletIndex() >= MaxNumberWalletsPerAccount {
		wallet, err = acc.GetInternalWallet(rand.UintN(MaxNumberWalletsPerAccount)) //nolint:gosec // no need for secure random here
	}
	if acc.GetInternalWalletIndex() < MaxNumberWalletsPerAccount {
		wallet, err = acc.GetNewInternalWallet()
	}

	if err != nil {
		return fmt.Errorf("error getting internal wallet: %w", err)
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
		return fmt.Errorf("error generating transaction (create and send): %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutSendTransaction)
	defer cancel()

	if errSend := acc.SendTransaction(ctx, tx); errSend != nil {
		return fmt.Errorf("error sending transaction: %w", errSend)
	}

	logger.Debugf("account %d distributed %f coins to %d addresses: %s",
		acc.GetAccountID(),
		kernel.ConvertFromChannoshisToCoins(util.GetBalanceUTXOs(utxos)),
		len(addresses),
		tx.String(),
	)

	return nil
}

// getRandomAccountAddress retrieves a random address for an account. If the limit of wallets per account has been
// reached, it will pick an existing wallet, otherwise it will create a new one
func getRandomAccountAddress(account *hd_wallet.Account) ([]byte, error) {
	var err error
	var wallet *wallt.Wallet

	// if the limit have been reached, pick an existing wallet
	if account.GetExternalWalletIndex() >= MaxNumberWalletsPerAccount {
		wallet, err = account.GetExternalWallet(rand.UintN(MaxNumberWalletsPerAccount)) //nolint:gosec // no need for secure random here
		if err != nil {
			return []byte{}, fmt.Errorf("error getting external wallet: %w", err)
		}
	}

	// if not all wallets have been created, create a new one
	if account.GetExternalWalletIndex() < MaxNumberWalletsPerAccount {
		wallet, err = account.GetNewExternalWallet()
		if err != nil {
			return []byte{}, fmt.Errorf("error getting new wallet: %w", err)
		}
	}

	address, errAddress := wallet.GetP2PKHAddress()
	if errAddress != nil {
		return []byte{}, fmt.Errorf("error getting P2PKH address: %w", errAddress)
	}

	return address, nil
}

// getRandomAccountAddresses retrieve a random number of addresses for an account between a minimum and maximum
func getRandomAccountAddresses(minRetrieve, maxRetrieve uint, account *hd_wallet.Account) ([][]byte, error) {
	var addresses [][]byte

	numAddresses := rand.UintN(maxRetrieve-minRetrieve) + minRetrieve //nolint:gosec // no need for secure random here

	for range numAddresses {
		address, errAddress := getRandomAccountAddress(account)
		if errAddress != nil {
			return [][]byte{}, fmt.Errorf("error getting random account address: %w", errAddress)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

func splitArrayRandomized[T any](array []T, maxLengthGroups int) [][]T {
	if maxLengthGroups <= 0 {
		logger.Fatalf("maxLengthGroups must be greater than 0")
	}

	// initialize a new random number generator with a seed based on the current time.
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewPCG(uint64(seed), 0)) //nolint:gosec // no need for secure random here

	var result [][]T
	for len(array) > 0 {
		// determine the size of the next group (random, up to maxLengthGroups)
		groupSize := rng.IntN(maxLengthGroups) + 1
		if groupSize > len(array) {
			groupSize = len(array)
		}

		// extract the group and append to the result
		result = append(result, array[:groupSize])

		// remove the group from the original array
		array = array[groupSize:]
	}

	return result
}
