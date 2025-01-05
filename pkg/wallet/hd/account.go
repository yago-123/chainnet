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
	wallt "github.com/yago-123/chainnet/pkg/wallet"
)

type HDAccount struct {
	// derivedPubAccountKey public key derived from the master private key to be used for this account
	// in other words, level 3 of the HD wallet derivation path
	derivedPubAccountKey []byte
	// derivedChainAccountCode chain code derived from the original master private key to be used for this account
	derivedChainAccountCode []byte
	// accountID represents the number that corresponds to this account (constant for each account)
	accountID uint32
	// walletNum represents the current index of the wallets generated via HD wallet
	walletNum uint32

	// todo(): maybe encapsulate these fields in a struct?
	walletVersion byte
	validator     consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer  sign.Signature
	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	cfg             *config.Config
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
	walletNum uint32,
) *HDAccount {
	return &HDAccount{
		cfg:                     cfg,
		walletVersion:           walletVersion,
		validator:               validator,
		signer:                  signer,
		consensusHasher:         consensusHasher,
		encoder:                 encoder,
		derivedPubAccountKey:    derivedPubAccountKey,
		derivedChainAccountCode: derivedChainAccountCode,
		accountID:               accountNum,
		walletNum:               walletNum, // this field will be 0 for new accounts and X during HD restoration
	}
}

func (hda *HDAccount) GetAccountID() uint32 {
	return hda.accountID
}

// NewWallet generates a new wallet based on the HD wallet derivation path. Although this method is called NewWallet,
// it should be called NewAddress according to BIP-44, but given that all the code is already written for a simple
// wallet, it's better to keep it this way for now and reuse the code related to wallet. Also have the advantage that
// it will isolate the network traces
func (hda *HDAccount) NewWallet() (*wallt.Wallet, error) {
	var err error
	var derivedPrivateKey []byte

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, in this case we are deriving
	// the account, so we only need the first three levels
	indexes := []uint32{
		uint32(ExternalChangeType), // given that the wallet does not have funds yet, the change type is external by default
		hda.walletNum,
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
		derivedPublicKey, err = util_crypto.DeriveECDSAPubFromPrivate(derivedPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
		}
	}

	// increment the wallet index and return the new wallet
	hda.walletNum++

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

func (hda *HDAccount) GetWalletIndex() uint32 {
	return hda.walletNum
}

func (hda *HDAccount) ConsolidateChange() {

}
