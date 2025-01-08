package hd

import (
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
	wallt "github.com/yago-123/chainnet/pkg/wallet"
)

type Account struct {
	// derivedPubAccountKey public key derived from the master private key to be used for this account
	// in other words, level 3 of the HD wallet derivation path
	derivedPubAccountKey []byte
	// derivedChainAccountCode chain code derived from the original master private key to be used for this account
	derivedChainAccountCode []byte
	// accountID represents the number that corresponds to this account (constant for each account)
	accountID uint32
	// walletNum represents the current index of the wallets generated via HD wallet
	walletNum uint32

	// mu mutex used to synchronize the access to the account fields
	mu sync.Mutex

	// todo(): maybe encapsulate these fields in a struct?
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

func NewHDAccount(
	cfg *config.Config,
	walletVersion byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	derivedPubAccountKey []byte,
	derivedChainAccountCode []byte,
	accountNum uint32,
) *Account {
	return &Account{
		cfg:                     cfg,
		logger:                  cfg.Logger,
		walletVersion:           walletVersion,
		validator:               validator,
		signer:                  signer,
		consensusHasher:         consensusHasher,
		encoder:                 encoder,
		derivedPubAccountKey:    derivedPubAccountKey,
		derivedChainAccountCode: derivedChainAccountCode,
		accountID:               accountNum,
	}
}

// Sync scans and updates the account by generating a series of wallets based on the default gap limit.
// It checks each wallet for funds by syncing with the network and looking for transactions. The gap limit
// defines the maximum number of consecutive empty wallets that can be generated before stopping the sync process.
// If a wallet contains transactions, it is considered active, and the process continues. If an account has
// no funds (empty wallet) for the specified gap limit, the syncing process halts, and the number of active wallets
// is recorded.
// Returns the number of active wallets found during the sync process and an error if any
func (hda *Account) Sync() (uint32, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	gaugeWalletsWithoutActivity := 0
	counterWalletsChecked := uint32(0)
	for {
		// generate wallet and check if had any activity (transactions)
		wallet, err := hda.createWallet(counterWalletsChecked)
		if err != nil {
			return 0, fmt.Errorf("error creating wallet: %w", err)
		}

		_, err = wallet.InitNetwork()
		if err != nil {
			return 0, fmt.Errorf("error setting up wallet network: %w", err)
		}

		txs, err := wallet.GetWalletTxs()
		if err != nil {
			return 0, fmt.Errorf("error getting wallet transactions: %w", err)
		}

		counterWalletsChecked++

		// if does not have funds increment the gauge
		if len(txs) == 0 {
			gaugeWalletsWithoutActivity++
		}

		// if does have funds reset the gauge
		if len(txs) > 0 {
			gaugeWalletsWithoutActivity = 0
		}

		// when the gauge is bigger than the gap limit, it means that we have reached the maximum number of consecutive
		// empty wallets, so we stop the syncing process
		if gaugeWalletsWithoutActivity >= AddressGapLimit {
			break
		}
	}

	// update the account number and return the number of active wallets found
	hda.walletNum = counterWalletsChecked - AddressGapLimit

	return hda.walletNum, nil
}

func (hda *Account) GetAccountID() uint32 {
	return hda.accountID
}

// GetNewWallet generates a new wallet based on the HD wallet derivation path
func (hda *Account) GetNewWallet() (*wallt.Wallet, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	wallet, err := hda.createWallet(hda.walletNum)
	if err != nil {
		return nil, fmt.Errorf("error creating new wallet: %w", err)
	}

	// increment the wallet index and return the new wallet
	hda.walletNum++
	return wallet, nil
}

func (hda *Account) GetWalletIndex() uint32 {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	return hda.walletNum
}

func (hda *Account) ConsolidateChange() {

}

// createWallet generates a new wallet by deriving by default the external and the wallet number selected. Although this
// method is called createWallet, it should be called createAddress according to BIP-44, but given that all the code
// is already written for a simple wallet, it's better to keep it this way for now and reuse the code related to wallet.
// Also have the advantage that it will isolate the network traces. This method does not persist the wallet in the
// Account object itself, it's the responsibility of the caller to do so if needed
func (hda *Account) createWallet(walletNum uint32) (*wallt.Wallet, error) {
	var err error
	var derivedPrivateKey []byte

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, in this case we are deriving
	// the account, so we only need the first three levels
	indexes := []uint32{
		uint32(ExternalChangeType), // given that the wallet does not have funds yet, the change type is external by default
		walletNum,
	}

	derivedPublicKey, derivedChainCode := hda.derivedPubAccountKey, hda.derivedChainAccountCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		// the derivedKey field is a public key, but the return value will be a private and chain code key
		derivedPrivateKey, derivedChainCode, err = DeriveChildStepNonHardened(derivedPublicKey, derivedChainCode, idx)
		if err != nil {
			return nil, fmt.Errorf("error deriving child key: %w", err)
		}

		// extract public key from derived key to be used for the subsequent non-hardened indexes OR the wallet creation
		derivedPublicKey, err = util_crypto.DeriveECDSAPubFromPrivateDERBytes(derivedPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
		}
	}

	// generate new wallet with the derived keys
	wallet, err := wallt.NewWalletWithKeys(
		hda.cfg,
		hda.walletVersion,
		hda.validator,
		hda.signer,
		hda.consensusHasher,
		hda.encoder,
		derivedPrivateKey,
		derivedPublicKey,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up wallet: %w", err)
	}

	return wallet, nil
}
