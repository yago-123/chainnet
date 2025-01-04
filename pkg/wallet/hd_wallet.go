package wallet

import (
	"fmt"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
)

// HDWallet represents a Hierarchical Deterministic wallet
type HDWallet struct {
	version    byte
	PrivateKey []byte // should be replaced by seed when BIP-39 is implemented

	masterPrivKey, masterPubKey []byte

	// masterChainCode is a 32-byte value used in the derivation of child keys. Prevents
	// an attacker from easily reconstructing the master private key from a child key or
	// from seeing the private key in the derivation process
	masterChainCode []byte

	accountIndex uint32
	changeIndex  uint32
	// walletIndex represents the current index of the wallets generated via HD wallet
	walletIndex uint32

	validator consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer sign.Signature

	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing

	cfg *config.Config
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
	publicKey []byte,
) (*HDWallet, error) {
	// this represents a variant of BIP-44 by skipping BIP-39
	masterInfo, err := util_crypto.CalculateHMACSha512([]byte(HMACKeyStandard), privateKey)
	if err != nil {
		return nil, fmt.Errorf("error calculating HMAC-SHA512 for master private key: %v", err)
	}

	masterPrivateKey := masterInfo[:32]
	masterChainCode := masterInfo[32:]
	masterPubKey, err := util_crypto.DeriveECDSAPubFromPrivate(masterPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error deriving public key from private key: %v", err)
	}

	// todo(): find out how many wallets are already created and update walletIndex
	return &HDWallet{
		cfg:             cfg,
		version:         version,
		PrivateKey:      privateKey,
		masterPrivKey:   masterPrivateKey,
		masterPubKey:    masterPubKey,
		masterChainCode: masterChainCode,
		walletIndex:     0,
		validator:       validator,
		signer:          signer,
		encoder:         encoder,
		consensusHasher: consensusHasher,
	}, nil
}

func (hd *HDWallet) CreateNewWallet() (*Wallet, error) {
	// todo(): add mutex?
	childPrivKey, _, err := hd.generateChildKey(HDPurposeBIP44, HDChainNetCoinType, 0, 0, hd.walletIndex)
	if err != nil {
		return nil, fmt.Errorf("error generating child key: %v", err)
	}
	hd.walletIndex++

	childPubKey, err := util_crypto.DeriveECDSAPubFromPrivate(childPrivKey)
	if err != nil {
		return nil, fmt.Errorf("error deriving public key from private key: %v", err)
	}

	return NewWalletWithKeys(
		hd.cfg,
		hd.version,
		hd.validator,
		hd.signer,
		hd.consensusHasher,
		hd.encoder,
		childPrivKey,
		childPubKey,
	)
}

func (hd *HDWallet) GetNewAccount() uint {
	return 0
}

func (hd *HDWallet) GetNewAddress(accountIndex uint) {

}

func (hd *HDWallet) GetAccount(accountIndex uint) {

}

func (hd *HDWallet) GetAddress(accountIndex uint, addressIndex uint) {

}

func (hd *HDWallet) ListAccounts() {

}

func (hd *HDWallet) reconstructHDWalletHistory() {

}

// generateChildKey generates a child key based on the provided arguments
func (hd *HDWallet) generateChildKey(purpose uint32, account uint32, change changeType, index uint32) ([]byte, []byte, error) {
	var err error

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, the last two are variable
	// and don't expose sensitive data
	indexes := []uint32{HardenedIndex + purpose, HardenedIndex + uint32(TypeChainNet), HardenedIndex + account, uint32(change), index}
	derivedPrivateKey, derivedChainCode := hd.masterPrivKey, hd.masterChainCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		// todo(): this implementation is not 100% correct, one of the reasons for hardened vs. non-hardened is the fact
		// todo(): that there is not a need of exposing the private key every time a child key is derived. In this case
		// todo(): we use the private key in all instances
		derivedPrivateKey, derivedChainCode, err = hd.deriveChildKey(derivedPrivateKey, derivedChainCode, idx)
		if err != nil {
			return nil, nil, err
		}
	}

	// return the final child private key and chain code
	return derivedPrivateKey, derivedChainCode, nil
}

// deriveChildKey derives a child key based on a master private key, master chain code and index
func (hd *HDWallet) deriveChildKey(privateKey []byte, chainCode []byte, index uint32) ([]byte, []byte, error) {
	// prepare the data for HMAC
	var data []byte

	// if corresponds to a hardened key, prepend 0x00 to the master private key. Hardened keys are more secure in theory
	// due to the fact that even the master pub key being compromised the child wallets are still secure
	if index >= HardenedIndex {
		// hardened key, prepend 0x00 to the master private key
		data = append([]byte{HardenedKeyPrefix}, privateKey...)
	}

	// if corresponds to a non-hardened key, prepend the master public key
	if index < HardenedIndex {
		// non-hardened key, prepend the master public key (public key is derived from private key)
		pubKey, err := util_crypto.DeriveECDSAPubFromPrivate(privateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("error deriving public key: %w", err)
		}

		// in this case we don't need to prepend 0x00 to the public key
		data = pubKey
	}
	// serialize index value as a 4-byte big-endian representation in byte array form
	data = append(data, byte(index>>24), byte(index>>16), byte(index>>8), byte(index))

	// apply initial hmac
	hmacOutput, err := util_crypto.CalculateHMACSha512(chainCode, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error calculating HMAC-SHA512 while deriving child key: %v", err)
	}

	childPrivateKey := hmacOutput[:32]
	childChainCode := hmacOutput[32:]

	// transform keys into big.Int to perform arithmetic operations
	childPrivateKeyInt := new(big.Int).SetBytes(childPrivateKey)
	masterPrivateKeyInt := new(big.Int).SetBytes(privateKey)

	// add the child key to the master key (mod curve order)
	childPrivateKeyInt.Add(childPrivateKeyInt, masterPrivateKeyInt)

	// ensure key is within valid range for elliptic curve operations
	curveOrder := btcec.S256().N
	childPrivateKeyInt.Mod(childPrivateKeyInt, curveOrder)
	// if the result is >= curve order, re-derive the key (this should not happen often)
	// todo(): add predefined error
	if childPrivateKeyInt.Cmp(curveOrder) >= 0 {
		// todo(): retrieve custom error
		return nil, nil, fmt.Errorf("child private key is invalid")
	}

	// return the child private key (as bytes) and child chain code
	return childPrivateKeyInt.Bytes(), childChainCode, nil
}
