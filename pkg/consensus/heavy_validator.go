package consensus

import "chainnet/pkg/kernel"

type HValidator struct {
	lv LightValidator
}

func NewHeavyValidator(lv LightValidator) *HValidator {
	return &HValidator{lv: lv}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) bool {
	if !hv.lv.ValidateTx(tx) {
		return false
	}

	// todo(): validate double spending check

	// todo(): validate timelock / block height constraints

	// todo(): maturity checks?

	return true
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) bool {
	// todo(): validate block size limit

	// todo(): validate block reward and fees

	return false
}
