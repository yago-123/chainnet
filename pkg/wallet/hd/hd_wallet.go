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

	accounts     []*HDAccount
	accountIndex uint32

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

func NewHDWallet() *HDWallet {
	return &HDWallet{}
}

func NewHDWalletWithKeys(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
	metadata *HDMetadata,
) (*HDWallet, error) {
	// this represents a variant of BIP-44 by skipping BIP-39
	masterInfo, err := util_crypto.CalculateHMACSha512([]byte(HMACKeyStandard), privateKey)
	if err != nil {
		return nil, fmt.Errorf("error calculating HMAC-SHA512 for master private key: %w", err)
	}

	masterPrivateKey := masterInfo[:32]
	masterChainCode := masterInfo[32:]
	masterPubKey, err := util_crypto.DeriveECDSAPubFromPrivate(masterPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
	}

	if metadata == nil {
		resyncHDFromNetwork()
	}

	if metadata != nil {
		resyncHDFromMetadata()
	}

	return &HDWallet{
		cfg:             cfg,
		walletVersion:   version,
		PrivateKey:      privateKey,
		masterPrivKey:   masterPrivateKey,
		masterPubKey:    masterPubKey,
		masterChainCode: masterChainCode,
		validator:       validator,
		signer:          signer,
		encoder:         encoder,
		consensusHasher: consensusHasher,
	}, nil
}

// NewAccount derives a new account from the HD wallet
func (hd *HDWallet) NewAccount() (uint, *HDAccount, error) {
	var err error

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, in this case we are deriving
	// the account, so we only need the first three levels
	indexes := []uint32{
		HardenedIndex | HDPurposeBIP44,
		HardenedIndex | uint32(TypeChainNet),
		HardenedIndex | hd.accountIndex,
	}

	derivedPrivateKey, derivedChainCode := hd.masterPrivKey, hd.masterChainCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		derivedPrivateKey, derivedChainCode, err = DeriveChildStepHardened(derivedPrivateKey, derivedChainCode, idx)
		if err != nil {
			return 0, nil, err
		}
	}

	hdAccount := NewHDAccount(
		hd.cfg,
		hd.walletVersion,
		hd.validator,
		hd.signer,
		hd.consensusHasher,
		hd.encoder,
		derivedPrivateKey,
		derivedChainCode,
		uint(hd.accountIndex),
	)
	hd.accounts = append(hd.accounts, hdAccount)
	hd.accountIndex++

	return uint(hd.accountIndex), hdAccount, nil
}

func (hd *HDWallet) GetAccount(accountIndex uint) (*HDAccount, error) {
	if uint32(accountIndex) >= hd.accountIndex {
		return nil, fmt.Errorf("account index %d does not exist", accountIndex)
	}

	return hd.accounts[accountIndex], nil
}

// GetMetadata returns the metadata of the HD wallet so that the state can be recovered without the need of resyncing
func (hd *HDWallet) GetMetadata() *HDMetadata {
	m := HDMetadata{}
	m.AccountIndex = hd.accountIndex

	for _, account := range hd.accounts {
		m.Accounts = append(m.Accounts, HDAccountMetadata{
			WalletIndex: account.GetWalletIndex(),
		})
	}

	return &m
}

func resyncHDFromNetwork() {

}

func resyncHDFromMetadata(metadata *HDMetadata) {
}
