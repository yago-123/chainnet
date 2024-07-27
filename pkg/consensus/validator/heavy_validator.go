package validator

import (
	"bytes"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"fmt"
)

type TxFunc func(tx *kernel.Transaction) error
type BlockFunc func(b *kernel.Block) error

type HValidator struct {
	lv       consensus.LightValidator
	explorer *explorer.Explorer
	signer   sign.Signature
	hasher   hash.Hashing
}

func NewHeavyValidator(lv consensus.LightValidator, explorer *explorer.Explorer, signer sign.Signature, hasher hash.Hashing) *HValidator {
	return &HValidator{
		lv:       lv,
		explorer: explorer,
		signer:   signer,
		hasher:   hasher,
	}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) error {
	validations := []TxFunc{
		hv.lv.ValidateTxLight,
		hv.validateOwnershipAndBalanceOfInputs,
		// todo(): validate timelock / block height constraints
		// todo(): maturity checks?
		// todo(): validate scriptSig
		// todo(): each input must have at least CPOMNASE_MATURITY(100) confirmations
	}

	for _, validate := range validations {
		if err := validate(tx); err != nil {
			return err
		}
	}

	return nil
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) error {
	validations := []BlockFunc{
		hv.validateBlockHash,
		hv.validateNumberOfCoinbaseTxs,
		hv.validateNoDoubleSpendingInsideBlock,
		// block header validations
		hv.validatePreviousBlockMatchCurrentLatest,
		hv.validateBlockHeight,
		hv.validateMerkleTree,
		hv.validateBlockTarget,
		// todo(): validate block size limit
		// todo(): validate coinbase transaction
		// todo(): validate block timestamp
	}

	for _, validate := range validations {
		if err := validate(b); err != nil {
			return err
		}
	}

	return nil
}

// validateOwnershipAndBalanceOfInputs checks that the inputs of a transaction are owned by the spender and that the
// balance of the inputs is greater or equal than the balance of the outputs
func (hv *HValidator) validateOwnershipAndBalanceOfInputs(tx *kernel.Transaction) error {
	// assume that we only use P2PK for now
	inputBalance := uint(0)
	outputBalance := uint(0)

	for _, vin := range tx.Vin {
		// fetch the unspent outputs for the input's public key
		utxos, _ := hv.explorer.FindUnspentOutputs(vin.PubKey)
		for _, utxo := range utxos {
			// if there is match, check that the signature is valid
			if utxo.EqualInput(vin) {
				// todo(): assume is P2PK only for now

				// check that the signature is valid for unlocking the UTXO
				sigCheck, err := hv.signer.Verify([]byte(vin.ScriptSig), tx.AssembleForSigning(), []byte(utxo.Output.PubKey))
				if err != nil {
					return fmt.Errorf("error verifying signature: %s", err.Error())
				}

				if !sigCheck {
					return fmt.Errorf("input with id %s and index %d has invalid signature", vin.Txid, vin.Vout)
				}

				// append the balance
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

// validateNumberOfCoinbaseTxs checks that there is only one coinbase transaction in a block. If there is more than
// one coinbase transaction it means that there has been an error adding multiple coinbases or that there are
// transactions with wrong number of inputs todo(): we may want to check this second case as well in the mempool
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

// validateBlockHash checks that the hash of the block is correct. Merkle tree hash is checked in validateMerkleTree func
func (hv *HValidator) validateBlockHash(b *kernel.Block) error {
	return util.VerifyBlockHash(b.Header, b.Hash, hv.hasher)
}

// validatePreviousBlockMatchCurrentLatest checks that the previous block hash of the block matches the latest block
func (hv *HValidator) validatePreviousBlockMatchCurrentLatest(b *kernel.Block) error {
	lastChainBlock, err := hv.explorer.GetLastBlock()
	if err != nil {
		return err
	}

	if !bytes.Equal(b.Header.PrevBlockHash, lastChainBlock.Hash) {
		return fmt.Errorf("previous hash %s points to block different than latest in the chain %s", string(b.Header.PrevBlockHash), string(lastChainBlock.Hash))
	}

	return nil
}

// validateBlockHeight checks that the block height matches the current chain height
func (hv *HValidator) validateBlockHeight(b *kernel.Block) error {
	lastChainBlock, err := hv.explorer.GetLastBlock()
	if err != nil {
		return err
	}

	if !(b.Header.Height == (lastChainBlock.Header.Height + 1)) {
		return fmt.Errorf("new block %s with height %d does not match current chain height %d", string(b.Hash), b.Header.Height, lastChainBlock.Header.Height)
	}

	return nil
}

// validateMerkleTree checks that the Merkle tree root hash of the block matches the Merkle root hash in the block header
func (hv *HValidator) validateMerkleTree(b *kernel.Block) error {
	merkletree, err := consensus.NewMerkleTreeFromTxs(b.Transactions, hv.hasher)
	if err != nil {
		return err
	}

	if !bytes.Equal(merkletree.RootHash(), b.Header.MerkleRoot) {
		return fmt.Errorf("block %s has invalid Merkle root", string(b.Hash))
	}

	return nil
}

// validateBlockTarget checks that the block hash corresponds to the target
func (hv *HValidator) validateBlockTarget(b *kernel.Block) error {
	if !util.IsFirstNBytesZero(b.Hash, b.Header.Target) {
		return fmt.Errorf("block %s has invalid target %d", string(b.Hash), b.Header.Target)
	}

	return nil
}
