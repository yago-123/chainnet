package hd

import (
	"fmt"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	cerror "github.com/yago-123/chainnet/pkg/error"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
)

// HDWallet represents a Hierarchical Deterministic wallet
type HDWallet struct {
	PrivateKey []byte // todo(): should be replaced by seed when BIP-39 is implemented

	// masterPrivKey and masterPubKey are the private and public keys used to derive child keys obtained from the seed
	masterPrivKey, masterPubKey []byte

	// masterChainCode is a 32-byte value used in the derivation of child keys. Prevents
	// an attacker from easily reconstructing the master private key from a child key or
	// from seeing the private key in the derivation process
	masterChainCode []byte

	accounts   []*HDAccount
	accountNum uint32

	// todo(): maybe encapsulate these fields in a struct?
	// walletVersion represents the version of the wallet being used
	walletVersion byte
	validator     consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer  sign.Signature
	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	cfg             *config.Config
}

func NewHDWalletWithKeys(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
) (*HDWallet, error) {
	var accountIndex uint32

	// todo(): enclose the master key derivation into a separate function
	// this represents a variant of BIP-44 by skipping BIP-39
	masterInfo, err := util_crypto.CalculateHMACSha512([]byte(HMACKeyStandard), privateKey)
	if err != nil {
		return nil, fmt.Errorf("error calculating HMAC-SHA512 for master private key: %w", err)
	}

	masterPrivateKey := masterInfo[:32]
	masterChainCode := masterInfo[32:]

	// the master private key is retrieved in raw format, convert to DER
	masterPrivateKeyDER, err := util_crypto.EncodeRawPrivateKeyToDERBytes(masterPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error encoding master private key to DER: %w", err)
	}

	// derive the public key from the master private key
	masterPubKey, err := util_crypto.DeriveECDSAPubFromPrivateDERBytes(masterPrivateKeyDER)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
	}

	return &HDWallet{
		cfg:             cfg,
		walletVersion:   version,
		PrivateKey:      privateKey,
		masterPrivKey:   masterPrivateKeyDER,
		masterPubKey:    masterPubKey,
		masterChainCode: masterChainCode,
		accountNum:      accountIndex,
		validator:       validator,
		signer:          signer,
		encoder:         encoder,
		consensusHasher: consensusHasher,
	}, nil
}

// Sync synchronizes the HD wallet fields so that all accounts and addresses are up to date
func (hd *HDWallet) Sync(metadata *HDMetadata) error {
	// if no metadata is provided, try to connect to the network and find which was the previous state
	if metadata == nil {
		return hd.resyncHDFromNetwork()
	}

	// if metadata is provided, try to resync the wallet from the metadata provided. Even if the metadata was wrong
	// this would not represent a danger to the funds, but could create privacy issues
	return hd.resyncHDFromMetadata(metadata)
}

// GetNewAccount derives a new account from the HD wallet by incrementing the account index
func (hd *HDWallet) GetNewAccount() (*HDAccount, error) {
	// we create a new
	_, account, err := hd.createAccount(hd.accountNum, 0)
	if err != nil {
		return nil, fmt.Errorf("error creating account: %w", err)
	}

	hd.accounts = append(hd.accounts, account)
	hd.accountNum++

	return account, nil
}

// GetAccount returns an account from the HD wallet by its index
func (hd *HDWallet) GetAccount(accountIdx uint) (*HDAccount, error) {
	if uint32(accountIdx) >= hd.accountNum {
		return nil, fmt.Errorf("account index %d does not exist", accountIdx)
	}

	return hd.accounts[accountIdx], nil
}

// GetMetadata returns the metadata of the HD wallet so that the state can be recovered without the need of resyncing
func (hd *HDWallet) GetMetadata() *HDMetadata {
	m := HDMetadata{}
	m.AccountNum = hd.accountNum

	for _, account := range hd.accounts {
		m.Accounts = append(m.Accounts, HDAccountMetadata{
			WalletNum: account.GetWalletIndex(),
		})
	}

	return &m
}

// createAccount creates a new account from the HD wallet with a given account number
// Arguments:
// - accountIndex: the index of the account to be created
// - walletNum: the number of wallets CREATED SO FAR! This is different from walletIndex. We use this field
// for restoring the HD wallet from the metadata, it does not affect in any way the key derivation in this part
// of the code. In the case of creating a new account, the walletNum will always be 0
func (hd *HDWallet) createAccount(accountIdx uint32, walletNum uint32) (uint32, *HDAccount, error) {
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
			return uint32(0), nil, err
		}
	}

	// derive the public key from the private key and pass paste it into the new account
	derivedPublicKey, err := util_crypto.DeriveECDSAPubFromPrivateDERBytes(derivedPrivateKey)
	if err != nil {
		return uint32(0), nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
	}

	hdAccount := NewHDAccount(
		hd.cfg,
		hd.walletVersion,
		hd.validator,
		hd.signer,
		hd.consensusHasher,
		hd.encoder,
		derivedPublicKey,
		derivedChainCode,
		accountIdx,
		walletNum,
	)

	return accountIdx, hdAccount, nil
}

// resyncHDFromNetwork resyncs the HD wallet from the network
func (hd *HDWallet) resyncHDFromNetwork() error {
	// todo() implement
	return nil
}

// resyncHDFromMetadata resyncs the HD wallet from the metadata
func (hd *HDWallet) resyncHDFromMetadata(metadata *HDMetadata) error {
	for accountIdx, accountMetadata := range metadata.Accounts {
		_, account, err := hd.createAccount(uint32(accountIdx), accountMetadata.WalletNum)
		if err != nil {
			return fmt.Errorf("error syncing account %d: %w", accountIdx, err)
		}

		hd.accounts = append(hd.accounts, account)
		hd.accountNum++
	}

	return nil
}
