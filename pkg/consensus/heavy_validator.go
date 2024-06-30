package consensus

import (
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
)

type HValidator struct {
	lv       LightValidator
	explorer explorer.Explorer
	signer   sign.Signature
}

func NewHeavyValidator(lv LightValidator, explorer explorer.Explorer, signer sign.Signature) *HValidator {
	return &HValidator{
		lv:       lv,
		explorer: explorer,
		signer:   signer,
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

	if !hv.validateOwnershipOfInputs(tx) {
		return false
	}

	// todo(): validate timelock / block height constraints

	// todo(): maturity checks?

	return true
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) bool {
	// todo(): validate hashes

	// todo(): validate block size limit

	if !hv.validateThereIsOnlyOneCoinbase(b) {
		return false
	}

	if !hv.validateNoDoubleSpendingInsideBlock(b) {
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
				break
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
				break
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

// validateOwnershipOfInputs checks that the inputs of a transaction are owned by the spender
func (hv *HValidator) validateOwnershipOfInputs(tx *kernel.Transaction) bool {
	// assume that we only use P2PK for now
	var err error

	for _, vin := range tx.Vin {
		validInput := false
		// fetch the unspent outputs for the input's public key
		utxos, _ := hv.explorer.FindUnspentOutputs(vin.PubKey)
		for _, utxo := range utxos {
			// if there is match, check that the signature is valid
			if utxo.EqualInput(vin) {
				// assume is P2PK only for now
				validInput, err = hv.signer.Verify([]byte(vin.ScriptSig), tx.AssembleForSigning(), []byte(utxo.Output.PubKey))
				if err != nil {
					return false
				}

				break
			}
		}

		if !validInput {
			return false
		}
	}

	return true
}

// validateNoDoubleSpendingInsideBlock checks that there are no repeated inputs inside a block
func (hv *HValidator) validateNoDoubleSpendingInsideBlock(b *kernel.Block) bool {
	// match every transaction with every other transaction
	for i := 0; i < len(b.Transactions); i++ {
		for j := i + 1; j < len(b.Transactions); j++ {

			// make sure that inputs do not match
			for _, vin := range b.Transactions[i].Vin {
				for _, vin2 := range b.Transactions[j].Vin {
					if vin.EqualInput(vin2) {
						return false
					}
				}
			}

		}
	}

	return true
}
