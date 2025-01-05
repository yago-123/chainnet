package util

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/kernel"
)

const (
	NumBitsInByte   = 8
	BiggestByteMask = 0xFF

	TargetAdjustmentUnit = uint(1)
	InitialBlockTarget   = uint(1)
	MinimumTarget        = uint(1)
	MaximumTarget        = uint(255)

	MinLengthHash = 16
	MaxLengthHash = 256
)

// CalculateTxHash calculates the hash of a transaction
func CalculateTxHash(tx *kernel.Transaction, hasher hash.Hashing) ([]byte, error) {
	// todo(): move this to the NewTransaction function instead?
	return hasher.Hash(tx.Assemble())
}

// VerifyTxHash verifies the hash of a transaction
func VerifyTxHash(tx *kernel.Transaction, hash []byte, hasher hash.Hashing) error {
	ret, err := hasher.Verify(hash, tx.Assemble())
	if err != nil {
		return fmt.Errorf("verify tx hash failed: %w", err)
	}

	if !ret {
		return errors.New("tx hash verification failed")
	}

	return nil
}

// CalculateBlockHash calculates the hash of a block header
func CalculateBlockHash(bh *kernel.BlockHeader, hasher hash.Hashing) ([]byte, error) {
	return hasher.Hash(bh.Assemble())
}

// VerifyBlockHash verifies the hash of a block header
func VerifyBlockHash(bh *kernel.BlockHeader, hash []byte, hasher hash.Hashing) error {
	ret, err := hasher.Verify(hash, bh.Assemble())
	if err != nil {
		return fmt.Errorf("block hashing failed: %w", err)
	}

	if !ret {
		return errors.New("block header hash verification failed")
	}

	return nil
}

// SafeUintToInt converts uint to int safely, returning an error if it would overflow
func SafeUintToInt(u uint) (int, error) {
	if u > uint(int(^uint(0)>>1)) { // Check if u exceeds the maximum int value
		return 0, errors.New("uint value exceeds int range")
	}

	return int(u), nil //nolint:gosec // This has been already checked
}

// IsFirstNBitsZero checks if the first n bits of the array are zero
func IsFirstNBitsZero(arr []byte, n uint) bool {
	if n == 0 {
		return true // if n is 0, trivially true
	}

	fullBytes, err := SafeUintToInt(n / NumBitsInByte)
	if err != nil {
		return false
	}
	remainingBits := n % NumBitsInByte

	arrLen := len(arr)
	if arrLen < fullBytes || (arrLen == fullBytes && remainingBits > 0) {
		return false
	}

	// check full bytes
	for i := range make([]struct{}, fullBytes) {
		if arr[i] != 0 {
			return false
		}
	}

	// check remaining bits in the next byte if there are any
	if remainingBits > 0 {
		nextByte := arr[fullBytes]
		mask := byte(BiggestByteMask << (NumBitsInByte - remainingBits))
		if nextByte&mask != 0 {
			return false
		}
	}

	return true
}

// CalculateMiningTarget calculates the new mining target based on the time required for mining the blocks
// vs. the time expected to mine the blocks:
//   - if required > expected -> decrease the target by 1 unit
//   - if required < expected -> increase the target by 1 unit
//   - if required = expected -> do not change the target
//
// The mechanism used is simplified to prevent high fluctuations.
func CalculateMiningTarget(currentTarget uint, targetTimeSpan float64, actualTimeSpan int64) uint {
	// determine the adjustment factor based on the actual and expected time spans
	timeAdjustmentFactor := float64(actualTimeSpan) / targetTimeSpan

	newTarget := currentTarget

	if timeAdjustmentFactor > 1.0 {
		// actual mining time is longer than expected, make it harder to mine
		newTarget = currentTarget - TargetAdjustmentUnit
	}

	if timeAdjustmentFactor < 1.0 {
		// actual mining time is shorter than expected, make it harder to mine
		newTarget = currentTarget + TargetAdjustmentUnit
	}

	// ensure the new target is within the valid range (there is no Min and Max for uint...)
	return uint(math.Min(math.Max(float64(newTarget), float64(MinimumTarget)), float64(MaximumTarget)))
}

func IsValidAddress(_ []byte) bool {
	// todo(): develop a proper address validation mechanism
	return true
}

func IsValidHash(hash []byte) bool {
	// convert raw []byte to a hexadecimal string
	hexString := fmt.Sprintf("%x", hash)

	// check length constraint
	if len(hexString) < MinLengthHash || len(hexString) > MaxLengthHash {
		return false
	}

	return true
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, fmt.Errorf("error opening file: %w", path, err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading file: %w", err)
	}

	return fileContent, nil
}
