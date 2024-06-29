package consensus

import (
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/kernel"
)

type HValidator struct {
	lv       LightValidator
	explorer explorer.Explorer
	// storage storage.Storage
}

func NewHeavyValidator(lv LightValidator, explorer explorer.Explorer) *HValidator {
	return &HValidator{
		lv:       lv,
		explorer: explorer,
	}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) bool {
	if !hv.lv.ValidateTx(tx) {
		return false
	}

	if !hv.validateInputRemainUnspent(tx) {
		return false
	}

	if !hv.validateBalance(tx) {
		return false
	}

	// todo(): validate that the input is a valid transaction
	// todo(): validate that the input is not already spent
	// todo(): validate that the input is owned by the spender
	// todo(): validate that the input is not a double spend

	// todo(): validate double spending check

	// todo(): validate inputs equal outputs balance

	// todo(): validate timelock / block height constraints

	// todo(): maturity checks?

	return true
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) bool {
	// todo(): validate hashes

	// todo(): validate there is not multiple transactions with same inputs

	// todo(): validate block size limit

	// todo(): validate block reward and fees

	// todo(): validate that there is only one coinbase transaction
	if !hv.validateThereIsOnlyOneCoinbase(b) {
		return false
	}

	return false
}

// validateThereIsOnlyOneCoinbase checks that there is only one coinbase transaction in a block
func (hv *HValidator) validateThereIsOnlyOneCoinbase(b *kernel.Block) bool {
	numCoinbases := 0
	for _, tx := range b.Transactions {
		if tx.IsCoinbase() {
			numCoinbases++
		}
	}

	return numCoinbases == 1
}

// validateInputRemainUnspent checks that the inputs of a transaction are not already spent by another transaction
func (hv *HValidator) validateInputRemainUnspent(tx *kernel.Transaction) bool {
	// skip the coinbase because does not have valid inputs
	if tx.IsCoinbase() {
		return true
	}

	// range over each input of the transaction
	for _, vin := range tx.Vin {
		validInput := false

		// fetch the unspent outputs for the input's public key
		utxos, _ := hv.explorer.FindUnspentOutputs(vin.PubKey)
		for _, utxo := range utxos {
			// if there is match, the input is valid and not spent
			if utxo.EqualInput(vin) {
				validInput = true
			}
		}

		// if there is not any utxo that match, the input is invalid
		if !validInput {
			return false
		}
	}

	return true
}

// validateBalance checks that the input balance is equal or lower than the output balance of a transaction
// todo() very similar to validateInputRemainUnspent, refactor eventually to avoid code duplication
func (hv *HValidator) validateBalance(tx *kernel.Transaction) bool {
	inputBalance := uint(0)
	outputBalance := uint(0)

	// retrieve the input balance
	for _, vin := range tx.Vin {
		// fetch the unspent outputs for the input's public key
		utxos, _ := hv.explorer.FindUnspentOutputs(vin.PubKey)
		for _, utxo := range utxos {
			// if there is match, retrieve the amount
			if utxo.EqualInput(vin) {
				inputBalance += utxo.Output.Amount
			}
		}
	}

	// retrieve the output balance
	for _, vout := range tx.Vout {
		outputBalance += vout.Amount
	}

	// make sure that the input balance is greater than the output balance (can be equal)
	return inputBalance < outputBalance
}
