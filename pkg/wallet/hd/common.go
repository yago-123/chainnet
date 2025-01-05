package hd

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	cerror "github.com/yago-123/chainnet/pkg/error"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
)

// General constants for HD wallet implementation
const (
	HMACKeyStandard           = "ChainNet seed"
	HardenedIndex      uint32 = 1 << 31
	HardenedKeyPrefix         = 0x00
	HDPurposeBIP44            = 44
	HDChainNetCoinType        = 0

	GapLimit = 1
)

type coinType uint32

const (
	TypeBitcoin               coinType = 0x1
	TypeTestnet                        = 0x2
	TypeLitecoin                       = 0x3
	TypeDogecoin                       = 0x4
	TypeReddcoin                       = 0x5
	TypeDash                           = 0x6
	TypePeercoin                       = 0x7
	TypeNamecoin                       = 0x8
	TypeFeathercoin                    = 0x9
	TypeCounterparty                   = 0xa
	TypeBlackcoin                      = 0xb
	TypeNuShares                       = 0xc
	TypeNuBits                         = 0xd
	TypeMazacoin                       = 0xe
	TypeViacoin                        = 0xf
	TypeClearingHouse                  = 0x10
	TypeRubycoin                       = 0x11
	TypeGroestlcoin                    = 0x12
	TypeDigitalcoin                    = 0x13
	TypeCannacoin                      = 0x14
	TypeDigiByte                       = 0x15
	TypeOpenAssets                     = 0x16
	TypeMonacoin                       = 0x17
	TypeClams                          = 0x18
	TypePrimecoin                      = 0x19
	TypeNeoscoin                       = 0x1a
	TypeJumbucks                       = 0x1b
	TypeziftrCOIN                      = 0x1c
	TypeVertcoin                       = 0x1d
	TypeNXT                            = 0x1e
	TypeBurst                          = 0x1f
	TypeMonetaryUnit                   = 0x20
	TypeZoom                           = 0x21
	TypeVpncoin                        = 0x22
	TypeCanadaeCoin                    = 0x23
	TypeShadowCash                     = 0x24
	TypeParkByte                       = 0x25
	TypePandacoin                      = 0x26
	TypeStartCOIN                      = 0x27
	TypeMOIN                           = 0x2D
	TypeArgentum                       = 0x31
	TypeGlobalCurrencyReserve          = 0x32
	TypeNovacoin                       = 0x33
	TypeAsiacoin                       = 0x34
	TypeBitcoindark                    = 0x35
	TypeDopecoin                       = 0x36
	TypeTemplecoin                     = 0x37
	TypeAIB                            = 0x38
	TypeEDRCoin                        = 0x39
	TypeSyscoin                        = 0x3a
	TypeSolarcoin                      = 0x3b
	TypeSmileycoin                     = 0x3c
	TypeEther                          = 0x3d
	TypeEtherClassic                   = 0x3e
	TypeOpenChain                      = 0x40
	TypeOKCash                         = 0x45
	TypeDogecoinDark                   = 0x4d
	TypeElectronicGulden               = 0x4e
	TypeClubCoin                       = 0x4f
	TypeRichCoin                       = 0x50
	TypePotcoin                        = 0x51
	TypeQuarkcoin                      = 0x52
	TypeTerracoin                      = 0x53
	TypeGridcoin                       = 0x54
	TypeAuroracoin                     = 0x55
	TypeIXCoin                         = 0x56
	TypeGulden                         = 0x57
	TypeBitBean                        = 0x58
	TypeBata                           = 0x59
	TypeMyriadcoin                     = 0x5a
	TypeBitSend                        = 0x5b
	TypeUnobtanium                     = 0x5c
	TypeMasterTrader                   = 0x5d
	TypeGoldBlocks                     = 0x5e
	TypeSaham                          = 0x5f
	TypeChronos                        = 0x60
	TypeUbiquoin                       = 0x61
	TypeEvotion                        = 0x62
	TypeSaveTheOcean                   = 0x63
	TypeBigUp                          = 0x64
	TypeGameCredits                    = 0x65
	TypeDollarcoins                    = 0x66
	TypeZayedcoin                      = 0x67
	TypeDubaicoin                      = 0x68
	TypeStratis                        = 0x69
	TypeShilling                       = 0x6a
	TypePiggyCoin                      = 0x76
	TypeMonero                         = 0x80
	TypeNavCoin                        = 0x82
	TypeFactomFactoids                 = 0x83
	TypeFactomEntryCredits             = 0x84
	TypeZcash                          = 0x85
	TypeLisk                           = 0x86
	TypeFactomIdentity                 = 0x119
	TypeChainNet                       = 0x120
)

type changeType uint32

const (
	ExternalChangeType changeType = iota // ExternalChangeType for addresses shared with others
	InternalChangeType                   // InternalChangeType for not shared/visible addresses
)

// DeriveChildStepHardened derives a child key based on a master private key, master chain code and index. The main
// difference between this function and DeriveChildStepNonHardened is that this function prepends a constant to the
// private key before derivation starts. Hardened keys are more secure in theory due to the fact that even the master
// pub key being compromised the child wallets are still secure. Based on BIP-44 standard the first 3 levels of
// the derivation path (purpose, coin type and account) are hardened keys due to security reasons
func DeriveChildStepHardened(privateKey []byte, chainCode []byte, index uint32) ([]byte, []byte, error) {
	// hardened key requires prepending a constant to the private key before derivation starts
	derivedKey := append([]byte{HardenedKeyPrefix}, privateKey...)
	return deriveChildStep(derivedKey, chainCode, index)
}

// DeriveChildStepNonHardened derives a child key based on a master public key, unlike DeriveChildStepHardened this func
// does not require prepending a constant to the private key before derivation starts. This is because non-hardened keys
// are less secure than hardened keys, but they are still secure. Based on BIP-44 standard the last 2 levels of the
// derivation path (change and index) are non-hardened keys (but is not a requirement)
func DeriveChildStepNonHardened(publicKey []byte, chainCode []byte, index uint32) ([]byte, []byte, error) {
	// non hardened key is just prepended and not modified
	return deriveChildStep(publicKey, chainCode, index)
}

// deriveChildStep derives a child key based on a master private key, master chain code and index
func deriveChildStep(derivedKey []byte, chainCode []byte, index uint32) ([]byte, []byte, error) {
	var data []byte

	// serialize index value as a 4-byte big-endian representation in byte array form
	data = append(derivedKey, byte(index>>24), byte(index>>16), byte(index>>8), byte(index))

	// apply initial hmac
	hmacOutput, err := util_crypto.CalculateHMACSha512(chainCode, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error calculating HMAC-SHA512 while deriving child key: %w", err)
	}

	childPrivateKey := hmacOutput[:32]
	childChainCode := hmacOutput[32:]

	// transform keys into big.Int to perform arithmetic operations
	childPrivateKeyInt := new(big.Int).SetBytes(childPrivateKey)
	parentPrivateKeyInt := new(big.Int).SetBytes(derivedKey)

	// add the child key to the master key (mod curve order)
	childPrivateKeyInt.Add(childPrivateKeyInt, parentPrivateKeyInt)

	// ensure key is within valid range for elliptic curve operations
	curveOrder := btcec.S256().N
	childPrivateKeyInt.Mod(childPrivateKeyInt, curveOrder)
	// if the result is >= curve order, re-derive the key (this should not happen often)
	if childPrivateKeyInt.Cmp(curveOrder) >= 0 {
		return nil, nil, fmt.Errorf("%w: key exceeds curve order", cerror.ErrWalletInvalidChildPrivateKey)
	}

	// return the child private key (as bytes) and child chain code
	return childPrivateKeyInt.Bytes(), childChainCode, nil
}

// isHardened checks if the index contains the hardened bit activated
func isHardened(index uint32) bool {
	return index&HardenedIndex != 0
}
