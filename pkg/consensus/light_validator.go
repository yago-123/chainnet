package consensus

import "chainnet/pkg/kernel"

type LValidator struct {
}

func NewLightValidator() *LValidator {
	return &LValidator{}
}

func (lv *LValidator) ValidateTx(tx *kernel.Transaction) bool {
	// validate that the hash
	return validateInputs(tx) && validateOutputs(tx) && validateBalance(tx)
}

func validateInputs(tx *kernel.Transaction) bool {
	// check that there is at least one input
	if len(tx.Vin) < 1 {
		return false
	}

	// todo(): check ownership of inputs
	return false
}

func validateOutputs(tx *kernel.Transaction) bool {
	if len(tx.Vout) < 1 {
		return false
	}

	// todo(): check that there is transaction fee

	return false
}

func validateBalance(tx *kernel.Transaction) bool {

	// todo(): check that inputs equal outputs balance (take into account transaction fee too)
	return false
}
