package hd

import (
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	wallt "github.com/yago-123/chainnet/pkg/wallet"
)

type HDAccount struct {
	// derivedPubAccountKey public key derived from the master private key to be used for this account
	// in other words, level 3 of the HD wallet derivation path
	derivedPubAccountKey []byte
	// derivedChainAccountCode chain code derived from the original master private key to be used for this account
	derivedChainAccountCode []byte
	// accountNum represents the number that corresponds to this account (constant for each account)
	accountNum uint
	// walletIndex represents the current index of the wallets generated via HD wallet
	walletIndex uint32

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
	accountNum uint,
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
		accountNum:              accountNum,
	}
}

func NewHDAccountFromMetadata() {

}

func (hda *HDAccount) GetAccountNum() uint {
	return hda.accountNum
}
func (hda *HDAccount) ConsolidateChange() {

}

func (hda *HDAccount) NewWallet() (*wallt.Wallet, error) {
	return nil, nil
}
