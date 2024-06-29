package consensus

import (
	"chainnet/pkg/kernel"
)

type LValidator struct {
}

func NewLightValidator() *LValidator {
	return &LValidator{}
}

func (lv *LValidator) ValidateTx(tx *kernel.Transaction) bool {
	// check that there is at least one input in non-coinbase transactions
	if !tx.HaveInputs() && !tx.IsCoinbase() {
		return false
	}

	// make sure that there are outputs in the transaction
	if !tx.HaveOutputs() {
		return false
	}

	// validate there are not multiple Vins with same source
	if !lv.validateInputsDontMatch(tx) {
		return false
	}

	// todo(): check ownership of inputs and validate signatures

	// todo(): check that there is transaction fee

	return true
}

// validateInputsDontMatch checks that the inputs don't match creating double spending problems
func (lv *LValidator) validateInputsDontMatch(tx *kernel.Transaction) bool {
	for i := 0; i < len(tx.Vin); i++ {
		for j := i + 1; j < len(tx.Vin); j++ {
			if tx.Vin[i].EqualInput(tx.Vin[j]) {
				return false
			}
		}
	}

	return true
}
