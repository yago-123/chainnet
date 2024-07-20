package validator

import (
	"chainnet/pkg/kernel"
	"errors"
	"fmt"
)

type LValidator struct {
}

func NewLightValidator() *LValidator {
	return &LValidator{}
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
		return errors.New("transaction has multiple inputs with the same source")
	}

	// todo(): set limit to the number of inputs and outputs

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