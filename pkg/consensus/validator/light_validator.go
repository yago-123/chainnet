package validator

import (
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"errors"
	"fmt"
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

	// validate there are not multiple Vins with same source
	if err := lv.validateInputsDontMatch(tx); err != nil {
		return err
	}

	// validate transaction hash match the transaction ID field
	if err := lv.validateTxID(tx); err != nil {
		return err
	}
	// todo(): set limit to the number of inputs and outputs

	// todo(): make sure that transaction size is within limits

	// todo(): make sure number of sigops is within limits

	return nil
}

func (lv *LValidator) ValidateHeader(bh *kernel.BlockHeader) error {
	validations := []HeaderFunc{
		lv.validateHeaderFieldsWithinLimits,
		lv.validateVersion,
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

// validateVersion makes sure that the header version is correct (to be developed)
func (lv *LValidator) validateVersion(_ *kernel.BlockHeader) error {
	return nil
}

// validateInputsDontMatch checks that the inputs don't match creating double spending problems
func (lv *LValidator) validateInputsDontMatch(tx *kernel.Transaction) error {
	for i := range len(tx.Vin) {
		for j := i + 1; j < len(tx.Vin); j++ {
			if tx.Vin[i].EqualInput(tx.Vin[j]) {
				return fmt.Errorf("transaction %s has multiple inputs with the same source", string(tx.ID))
			}
		}
	}

	return nil
}

// validateTxID checks that the hash of the transaction matches the ID field
func (lv *LValidator) validateTxID(tx *kernel.Transaction) error {
	return util.VerifyTxHash(tx, tx.ID, lv.hasher)
}
