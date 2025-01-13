package hd_wallet

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	cerror "github.com/yago-123/chainnet/pkg/errs"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
)

const (
	MaxConcurrentRequests = 5
)

// Wallet represents a Hierarchical Deterministic wallet
type Wallet struct {
	PrivateKey []byte // todo(): should be replaced by seed when BIP-39 is implemented

	// masterPrivKey and masterPubKey are the private and public keys used to derive child keys obtained from the seed
	masterPrivKey, masterPubKey []byte

	// masterChainCode is a 32-byte value used in the derivation of child keys. Prevents
	// an attacker from easily reconstructing the master private key from a child key or
	// from seeing the private key in the derivation process
	masterChainCode []byte

	accounts []*Account

	// mu mutex used to synchronize the access to the HD wallet fields
	mu sync.Mutex

	// todo(): maybe encapsulate these fields in a struct?
	// walletVersion represents the version of the wallet being used
	walletVersion byte
	validator     consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer  sign.Signature
	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing

	logger *logrus.Logger
	cfg    *config.Config
}

func NewHDWalletWithKeys(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
) (*Wallet, error) {

	// todo(): enclose the master key derivation into a separate function
	// this represents a variant of BIP-44 by skipping BIP-39
	masterInfo, err := util_crypto.CalculateHMACSha512([]byte(HMACKeyStandard), privateKey)
	if err != nil {
		return nil, fmt.Errorf("error calculating HMAC-SHA512 for master private key: %w", err)
	}

	masterPrivateKey := masterInfo[:32]
	masterChainCode := masterInfo[32:]

	// the master private key is retrieved in raw format, convert to DER
	masterPrivateKeyDER, err := util_crypto.EncodeRawECDSAP256PrivateKeyToDERBytes(masterPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error encoding master private key to DER: %w", err)
	}

	// derive the public key from the master private key
	masterPubKey, err := util_crypto.DeriveECDSAPubFromPrivateDERBytes(masterPrivateKeyDER)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
	}

	return &Wallet{
		cfg:             cfg,
		logger:          cfg.Logger,
		walletVersion:   version,
		PrivateKey:      privateKey,
		masterPrivKey:   masterPrivateKeyDER,
		masterPubKey:    masterPubKey,
		masterChainCode: masterChainCode,
		validator:       validator,
		signer:          signer,
		encoder:         encoder,
		consensusHasher: consensusHasher,
	}, nil
}

func NewHDWalletWithMetadata(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
	metadata *Metadata) (*Wallet, error) {

	// create base HD wallet
	hdWallet, err := NewHDWalletWithKeys(cfg, version, validator, signer, consensusHasher, encoder, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating HD wallet with keys: %w", err)
	}

	// create accounts with the corresponding metadata
	for i := 0; i < len(metadata.MetadataAccounts); i++ {
		accountMetadata := metadata.MetadataAccounts[i]
		account, errAcc := hdWallet.createAccount(uint32(i), uint32(accountMetadata.NumExternalWallets), uint32(accountMetadata.NumInternalWallets))
		if errAcc != nil {
			return nil, fmt.Errorf("error creating account %d: %w", i, errAcc)
		}

		hdWallet.accounts = append(hdWallet.accounts, account)
	}

	return hdWallet, nil
}

// Sync synchronizes the HD wallet fields so that all accounts and addresses are up to date
func (hd *Wallet) Sync() (uint, error) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	tmpAccounts := []*Account{}
	accountsWithoutActivity := 0
	for {
		// generate accounts and check if had any activity (transactions)
		account, err := hd.createAccount(uint32(len(tmpAccounts)), 0, 0) //nolint:gosec // possibility of integer overflow is OK here
		if err != nil {
			return 0, fmt.Errorf("error creating account %d: %w", len(tmpAccounts), err)
		}
		tmpAccounts = append(tmpAccounts, account)

		// sync the account with the network
		numExternalWallets, numInternalWallets, err := account.Sync()
		if err != nil {
			return 0, fmt.Errorf("error syncing account %d: %w", account.GetAccountID(), err)
		}

		// if the account has no addresses in use, increment the gauge. In theory, if the external account is 0, the
		// internal must be 0 as well, but just in case we perform the check too
		if numExternalWallets == 0 && numInternalWallets == 0 {
			accountsWithoutActivity++
		}

		// if the account has addresses in use, reset the gauge
		if numExternalWallets > 0 || numInternalWallets > 0 {
			hd.logger.Debugf("synced account %d: have %d external and %d internal accounts ", account.GetAccountID(), numExternalWallets, numInternalWallets)
			accountsWithoutActivity = 0
		}

		// if the gap limit is reached, stop the sync process
		if accountsWithoutActivity >= AccountGapLimit {
			break
		}
	}

	// store the accounts that are in use and update the account number
	hd.accounts = tmpAccounts[:len(tmpAccounts)-AccountGapLimit]

	return uint(len(hd.accounts)), nil
}

func (hd *Wallet) GetMetadata() *Metadata {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	metadata := Metadata{}

	metadata.NumAccounts = uint(len(hd.accounts))

	for _, account := range hd.accounts {
		metadata.MetadataAccounts = append(metadata.MetadataAccounts, MetadataAccount{
			NumExternalWallets: account.GetExternalWalletIndex(),
			NumInternalWallets: account.GetInternalWalletIndex(),
		})
	}

	return &metadata
}

// GetNewAccount derives a new account from the HD wallet by incrementing the account index
func (hd *Wallet) GetNewAccount() (*Account, error) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	// we create a new account and increment the account number
	account, err := hd.createAccount(uint32(len(hd.accounts)), 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error creating account: %w", err)
	}

	hd.accounts = append(hd.accounts, account)

	return account, nil
}

// GetAccount returns an account from the HD wallet by its index
func (hd *Wallet) GetAccount(accountIdx uint) (*Account, error) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	if accountIdx >= uint(len(hd.accounts)) { ///nolint:gosec // possibility of integer overflow is OK here
		return nil, fmt.Errorf("account index %d does not exist", accountIdx)
	}

	return hd.accounts[accountIdx], nil
}

func (hd *Wallet) GetNumAccounts() uint {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	return uint(len(hd.accounts))
}

// GetBalance returns the total balance of the HD wallet by summing the balances of all accounts. In case of finding
// one error, the method will return the error and stop processing the other accounts. This method ensures that the
// balance reported is the most accurate possible (hence why we cancel and return error if we find 1 error)
func (hd *Wallet) GetBalance() (uint, error) {
	// create a cancellable context, this will be used in case an error is found (everything will be canceled)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure cancel is called to release resources

	semaphore := make(chan struct{}, MaxConcurrentRequests) // semaphore to limit concurrency
	var wg sync.WaitGroup
	var balanceMu sync.Mutex // protects totalBalance

	totalBalance := uint(0)

	var overallErr error
	var overallErrMu sync.Mutex // protects overallErr

	for _, account := range hd.accounts {
		semaphore <- struct{}{} // acquire a slot
		wg.Add(1)

		go func(acc *Account) {
			defer wg.Done()
			defer func() { <-semaphore }() // release the slot

			select {
			case <-ctx.Done():
				// context canceled, stop processing
				return
			default:
				// proceed with fetching balance
				balance, err := acc.GetBalance()
				if err != nil {
					// capture the error and cancel other operations
					overallErrMu.Lock()
					if overallErr == nil { // only set the first error
						overallErr = fmt.Errorf("error getting balance for account %d: %w", acc.GetAccountID(), err)
						cancel() // cancel the context
					}
					overallErrMu.Unlock()
					return
				}

				// safely add the account's balance to the total balance
				balanceMu.Lock()
				totalBalance += balance
				balanceMu.Unlock()
			}
		}(account)
	}

	wg.Wait() // wait for all goroutines to finish

	if overallErr != nil {
		return 0, overallErr
	}

	return totalBalance, nil
}

// createAccount creates an  account from the HD wallet with a given account number. This method does not persist the
// account in the HD wallet, it's the responsibility of the caller to do so if needed.
// Arguments:
// - accountIndex: the index of the account to be created
func (hd *Wallet) createAccount(accountIdx uint32, externalWalletsIdx, internalWalletsIdx uint32) (*Account, error) {
	var err error

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, in this case we are deriving
	// the account, so we only need the first three levels
	indexes := []uint32{
		HardenedIndex | HDPurposeBIP44,
		HardenedIndex | uint32(TypeChainNet),
		HardenedIndex | accountIdx,
	}

	derivedPrivateKey, derivedChainCode := hd.masterPrivKey, hd.masterChainCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		derivedPrivateKey, derivedChainCode, err = DeriveChildStepHardened(derivedPrivateKey, derivedChainCode, idx)
		if err != nil {
			return nil, err
		}
	}

	// derive the public key from the private key and pass paste it into the new account
	derivedPublicKey, err := util_crypto.DeriveECDSAPubFromPrivateDERBytes(derivedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
	}

	return NewHDAccount(
		hd.cfg,
		hd.walletVersion,
		hd.validator,
		hd.signer,
		hd.consensusHasher,
		hd.encoder,
		derivedPublicKey,
		derivedChainCode,
		accountIdx,
		externalWalletsIdx,
		internalWalletsIdx,
	)
}
