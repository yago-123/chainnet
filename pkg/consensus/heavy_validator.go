package consensus

import (
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"fmt"
)

type ValidatorTxFunc func(tx *kernel.Transaction) error
type ValidatorBlockFunc func(b *kernel.Block) error

type HValidator struct {
	lv       LightValidator
	explorer explorer.Explorer
	signer   sign.Signature
	hasher   hash.Hashing
}

func NewHeavyValidator(lv LightValidator, explorer explorer.Explorer, signer sign.Signature, hasher hash.Hashing) *HValidator {
	return &HValidator{
		lv:       lv,
		explorer: explorer,
		signer:   signer,
		hasher:   hasher,
	}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) error {
	validations := []ValidatorTxFunc{
		hv.lv.ValidateTxLight,
		hv.validateInputRemainUnspent,
		hv.validateBalance,
		hv.validateOwnershipOfInputs,
		// todo(): validate timelock / block height constraints
		// todo(): maturity checks?
	}

	for _, validate := range validations {
		if err := validate(tx); err != nil {
			return err
		}
	}

	return nil
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) error {
	validations := []ValidatorBlockFunc{
		hv.validateBlockHash,
		hv.validateNumberOfCoinbaseTxs,
		hv.validateNoDoubleSpendingInsideBlock,
		// todo(): validate block size limit
	}

	for _, validate := range validations {
		if err := validate(b); err != nil {
			return err
		}
	}

	return nil
}

// validateInputRemainUnspent checks that the inputs of a transaction are not already spent by another transaction
func (hv *HValidator) validateInputRemainUnspent(tx *kernel.Transaction) error {
	// skip the coinbase because does not have valid inputs
	if tx.IsCoinbase() {
		return nil
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
			return fmt.Errorf("input with id %s and index %d is already spent", vin.Txid, vin.Vout)
		}
	}

	return nil
}

// validateBalance checks that the input balance is equal or lower than the output balance of a transaction
// todo() very similar to validateInputRemainUnspent, refactor eventually to avoid code duplication
func (hv *HValidator) validateBalance(tx *kernel.Transaction) error {
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
	if inputBalance > outputBalance {
		return fmt.Errorf("input balance %d is greater than output balance %d", inputBalance, outputBalance)
	}

	return nil
}

// validateOwnershipOfInputs checks that the inputs of a transaction are owned by the spender
func (hv *HValidator) validateOwnershipOfInputs(tx *kernel.Transaction) error {
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
					return fmt.Errorf("input with id %s and index %d has invalid signature", vin.Txid, vin.Vout)
				}

				break
			}
		}

		if !validInput {
			return fmt.Errorf("input with id %s and index %d have been already spent", vin.Txid, vin.Vout)
		}
	}

	return nil
}

// validateNumberOfCoinbaseTxs checks that there is only one coinbase transaction in a block
func (hv *HValidator) validateNumberOfCoinbaseTxs(b *kernel.Block) error {
	numCoinbases := 0
	for _, tx := range b.Transactions {
		if tx.IsCoinbase() {
			numCoinbases++
		}
	}

	if numCoinbases == 0 {
		return fmt.Errorf("block %s has no coinbase transaction", string(b.Hash))
	}

	if numCoinbases > 1 {
		return fmt.Errorf("block %s has more than one coinbase transaction", string(b.Hash))
	}

	return nil
}

// validateNoDoubleSpendingInsideBlock checks that there are no repeated inputs inside a block
func (hv *HValidator) validateNoDoubleSpendingInsideBlock(b *kernel.Block) error {
	// match every transaction with every other transaction
	for i := range len(b.Transactions) {
		for j := i + 1; j < len(b.Transactions); j++ {
			// make sure that inputs do not match
			for _, vin := range b.Transactions[i].Vin {
				for _, vin2 := range b.Transactions[j].Vin {
					if vin.EqualInput(vin2) {
						return fmt.Errorf("transaction %s has input that is also spent in transaction %s", string(b.Transactions[i].ID), vin.Txid)
					}
				}
			}

		}
	}

	return nil
}

func (hv *HValidator) validateBlockHash(b *kernel.Block) error {
	// todo() once we have Merkle tree

	return nil
}
