package consensus

import (
	"chainnet/pkg/kernel"
)

type HValidator struct {
	lv LightValidator
	// storage storage.Storage
}

func NewHeavyValidator(lv LightValidator) *HValidator {
	return &HValidator{lv: lv}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) bool {
	return !hv.lv.ValidateTx(tx)

	// todo(): validate double spending check

	// todo(): validate timelock / block height constraints

	// todo(): maturity checks?

}

func (hv *HValidator) ValidateBlock(b *kernel.Block) bool {
	// todo(): validate hashes

	// todo(): validate block size limit

	// todo(): validate block reward and fees

	return false
}
