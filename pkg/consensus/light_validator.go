package consensus

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
	// check that there is at least one input in non-coinbase transactions
	if !tx.HaveInputs() && !tx.IsCoinbase() {
		return errors.New("transaction has no inputs")
	}

	// make sure that there are outputs in the transaction
	if !tx.HaveOutputs() {
		return errors.New("transaction has no outputs")
	}

	// validate there are not multiple Vins with same source
	if lv.validateInputsDontMatch(tx) != nil {
		return errors.New("transaction has multiple inputs with the same source")
	}

	// todo(): check ownership of inputs and validate signatures

	// todo(): check that there is transaction fee

	return nil
}

// validateInputsDontMatch checks that the inputs don't match creating double spending problems
func (lv *LValidator) validateInputsDontMatch(tx *kernel.Transaction) error {
	for i := 0; i < len(tx.Vin); i++ {
		for j := i + 1; j < len(tx.Vin); j++ {
			if tx.Vin[i].EqualInput(tx.Vin[j]) {

				return fmt.Errorf("transaction %s has multiple inputs with the same source", string(tx.ID))
			}
		}
	}

	return nil
}
