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

const (
	HMACKeyStandard = "ChainNet seed"
	HardenedIndex   = 0x80000000
)

// HDWallet represents a Hierarchical Deterministic wallet
type HDWallet struct {
	version    byte
	PrivateKey []byte // should be replaced by seed when BIP-39 is implemented

	masterPrivKey, masterPubKey, masterChainCode []byte

	validator consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer sign.Signature

	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing

	cfg *config.Config
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

	return &HDWallet{
		cfg:             cfg,
		version:         version,
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

// deriveChildKey derives a child key based on a master private key, master chain code and index
func (hd *HDWallet) deriveChildKey(privateKey []byte, chainCode []byte, index uint32) ([]byte, []byte, error) {
	// prepare the data for HMAC
	var data []byte
	if index >= HardenedIndex {
		// hardened key, prepend 0x00 to the master private key
		data = append([]byte{0x00}, privateKey...)
	} else {
		// non-hardened key, prepend the master public key (public key is derived from private key)
		privKey, err := util_crypto.ConvertBytesToECDSAPriv(privateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("error converting private key: %w", err)
		}

		pubKey, err := util_crypto.ConvertECDSAPubToBytes(&privKey.PublicKey)
		if err != nil {
			return nil, nil, fmt.Errorf("error deriving public key: %w", err)
		}
		data = append([]byte{0x00}, pubKey...)
	}
	// append index (big-endian)
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
	if childPrivateKeyInt.Cmp(curveOrder) >= 0 {
		// todo(): retrieve custom error
		return nil, nil, fmt.Errorf("child private key is invalid")
	}

	// return the child private key (as bytes) and child chain code
	return childPrivateKeyInt.Bytes(), childChainCode, nil
}

// GenerateChildKey generates a child key based on the provided arguments
func (hd *HDWallet) GenerateChildKey(purpose uint32, coinType uint32, account uint32, change uint32, index uint32) ([]byte, []byte, error) {
	var err error

	// derive the child key step by step, following the BIP44 path
	indexes := []uint32{purpose, coinType, account, change, index}
	derivedPrivateKey, derivedChainCode := hd.masterPrivKey, hd.masterChainCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		derivedPrivateKey, derivedChainCode, err = hd.deriveChildKey(derivedPrivateKey, derivedChainCode, idx+HardenedIndex)
		if err != nil {
			return nil, nil, err
		}
	}

	// return the final child private key and chain code
	return derivedPrivateKey, derivedChainCode, nil
}
