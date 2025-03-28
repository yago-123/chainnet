package validator

import (
	"errors"
	"fmt"

	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/util"
)

type LValidator struct {
	hasher hash.Hashing
}

func NewLightValidator(hasher hash.Hashing) *LValidator {
	return &LValidator{
		hasher: hasher,
	}
}

func (lv *LValidator) ValidateTxLight(tx *kernel.Transaction) error {
	// check that there is at least one input
	if !tx.HaveInputs() {
		return errors.New("transaction has no inputs")
	}

	// make sure that there are outputs in the transaction
	if !tx.HaveOutputs() {
		return errors.New("transaction has no outputs")
	}

	validations := []TxFunc{
		lv.validateInputsDontMatch,
		lv.validateTxID,
		lv.validateAllOutputsContainNonZeroAmounts,
		// todo(): set limit to the number of inputs and outputs
		// todo(): make sure that transaction size is within limits
		// todo(): make sure number of sigops is within limits
	}

	for _, validate := range validations {
		if err := validate(tx); err != nil {
			return err
		}
	}

	return nil
}

func (lv *LValidator) ValidateHeader(bh *kernel.BlockHeader) error {
	validations := []HeaderFunc{
		lv.validateHeaderFieldsWithinLimits,
		lv.validateVersion,
		lv.validateHeaderHash,
	}

	for _, validate := range validations {
		if err := validate(bh); err != nil {
			return err
		}
	}

	return nil
}

// validateHeaderFieldsWithinLimits makes sure that the fields of the block header contain correct values
func (lv *LValidator) validateHeaderFieldsWithinLimits(bh *kernel.BlockHeader) error {
	if bh.Height > 0 && len(bh.PrevBlockHash) == 0 {
		return fmt.Errorf("previous block hash is empty")
	}

	if len(bh.MerkleRoot) == 0 {
		return fmt.Errorf("merkle root is empty")
	}

	return nil
}

// validateHeaderHash checks that the block hash corresponds to the target
func (lv *LValidator) validateHeaderHash(bh *kernel.BlockHeader) error {
	headerHash, err := util.CalculateBlockHash(bh, lv.hasher)
	if err != nil {
		return fmt.Errorf("error calculating header headerHash: %w", err)
	}

	if !util.IsFirstNBitsZero(headerHash, bh.Target) {
		return fmt.Errorf("block %x has invalid target %d", headerHash, bh.Target)
	}

	return nil
}

// validateVersion makes sure that the header version is correct (to be developed)
func (lv *LValidator) validateVersion(_ *kernel.BlockHeader) error {
	return nil
}

// validateInputsDontMatch checks that the inputs don't match creating double spending problems
func (lv *LValidator) validateInputsDontMatch(tx *kernel.Transaction) error {
	for i := range len(tx.Vin) {
		for j := i + 1; j < len(tx.Vin); j++ {
			if tx.Vin[i].EqualInput(tx.Vin[j]) {
				return fmt.Errorf("transaction %x has multiple inputs with the same source", tx.ID)
			}
		}
	}

	return nil
}

// validateTxID checks that the hash of the transaction matches the ID field
func (lv *LValidator) validateTxID(tx *kernel.Transaction) error {
	return util.VerifyTxHash(tx, tx.ID, lv.hasher)
}

// validateAllOutputsContainNonZeroAmounts make sure that no empty outputs can be accepted
func (lv *LValidator) validateAllOutputsContainNonZeroAmounts(tx *kernel.Transaction) error {
	for i, out := range tx.Vout {
		if out.Amount == 0 {
			return fmt.Errorf("transaction %x contain output %d empty", tx.ID, i)
		}
	}

	return nil
}
