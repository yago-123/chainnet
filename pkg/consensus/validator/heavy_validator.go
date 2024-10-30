package validator

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/yago-123/chainnet/pkg/storage"

	"github.com/yago-123/chainnet/config"

	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/util"
)

type TxFunc func(tx *kernel.Transaction) error
type HeaderFunc func(bh *kernel.BlockHeader) error
type BlockFunc func(b *kernel.Block) error

type HValidator struct {
	lv       consensus.LightValidator
	explorer *explorer.ChainExplorer
	signer   sign.Signature
	hasher   hash.Hashing
	cfg      *config.Config
}

func NewHeavyValidator(
	cfg *config.Config,
	lv consensus.LightValidator,
	explorer *explorer.ChainExplorer,
	signer sign.Signature,
	hasher hash.Hashing,
) *HValidator {
	return &HValidator{
		lv:       lv,
		explorer: explorer,
		signer:   signer,
		hasher:   hasher,
		cfg:      cfg,
	}
}

func (hv *HValidator) ValidateTx(tx *kernel.Transaction) error {
	validations := []TxFunc{
		hv.lv.ValidateTxLight,
		hv.validateOwnershipAndBalanceOfInputs,
		// todo(): validate timelock / block height constraints
		// todo(): maturity checks?
		// todo(): each input must have at least CPOMNASE_MATURITY(100) confirmations
	}

	for _, validate := range validations {
		if err := validate(tx); err != nil {
			return err
		}
	}

	return nil
}

func (hv *HValidator) ValidateHeader(bh *kernel.BlockHeader) error {
	validations := []HeaderFunc{
		hv.lv.ValidateHeader,
		hv.validateGenesisHeader,
		hv.validateHeaderHeight,
		hv.validateHeaderTarget,
		hv.validateHeaderPreviousBlock,
	}

	for _, validate := range validations {
		if err := validate(bh); err != nil {
			return err
		}
	}

	return hv.lv.ValidateHeader(bh)
}

func (hv *HValidator) ValidateBlock(b *kernel.Block) error {
	if err := hv.ValidateHeader(b.Header); err != nil {
		return fmt.Errorf("error validating block header: %w", err)
	}

	// todo() where the fuck is the block hash target validated???

	validations := []BlockFunc{
		hv.validateBlockHash,
		hv.ValidateBlockWithoutHash,

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

// ValidateBlockWithoutHash is a special case of ValidateBlock that does not check the block hash. This is useful for
// the miner to validate the block before mining it. It's used to ensure that the block is valid before mining it
func (hv *HValidator) ValidateBlockWithoutHash(b *kernel.Block) error {
	if err := hv.ValidateHeader(b.Header); err != nil {
		return fmt.Errorf("error validating block header: %w", err)
	}

	validations := []BlockFunc{
		hv.validateNumberOfCoinbaseTxs,
		hv.validateNoDoubleSpendingInsideBlock,
		hv.validateMerkleTree,

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
		// todo(): would make sense to add a check via UTXO set?
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

// validateHeaderPreviousBlock checks that the previous block hash of the block matches the latest block
func (hv *HValidator) validateHeaderPreviousBlock(bh *kernel.BlockHeader) error {
	// if is genesis block and does not contain previous block hash, don't check previous block (does not exist)
	if bh.IsGenesisHeader() {
		return nil
	}

	// if is height 0 but contains previous block hash, return error
	if bh.Height == 0 && len(bh.PrevBlockHash) != 0 {
		return fmt.Errorf("expected genesis block with height 0, but contains previous block hash %x", bh.PrevBlockHash)
	}

	// if not genesis block, check previous block hash
	lastChainHeader, err := hv.explorer.GetLastHeader()
	if err != nil {
		return fmt.Errorf("unable to retrieve last header: %w", err)
	}

	lastHeaderHash, err := util.CalculateBlockHash(lastChainHeader, hv.hasher)
	if err != nil {
		return fmt.Errorf("error while calculating hash of last header: %s", lastChainHeader.String())
	}

	if !bytes.Equal(bh.PrevBlockHash, lastHeaderHash) {
		return fmt.Errorf("previous hash %x points to block different than latest in the chain %x", bh.PrevBlockHash, lastHeaderHash)
	}

	return nil
}

// validateGenesisHeader checks that the genesis block is valid
func (hv *HValidator) validateGenesisHeader(bh *kernel.BlockHeader) error {
	if !bh.IsGenesisHeader() {
		return nil
	}

	// if is genesis block, check that there is not any existent header
	_, err := hv.explorer.GetLastHeader()
	if !errors.Is(err, storage.ErrNotFound) {
		return fmt.Errorf("genesis block already exists")
	}

	// if the error is storage.ErrNotFound, then the genesis block is valid
	return nil
}

// validateHeaderHeight checks that the height of the block is correct
func (hv *HValidator) validateHeaderHeight(bh *kernel.BlockHeader) error {
	if bh.IsGenesisHeader() {
		return nil
	}

	// if not genesis block, check previous block header height
	lastChainHeader, err := hv.explorer.GetLastHeader()
	if err != nil {
		return fmt.Errorf("unable to retrieve last header: %w", err)
	}

	if !(bh.Height == (lastChainHeader.Height + 1)) {
		return fmt.Errorf("header does not match local height")
	}

	return nil
}

// validateHeaderTarget checks that the target of the header is correct
func (hv *HValidator) validateHeaderTarget(bh *kernel.BlockHeader) error {
	targetExpected, err := hv.explorer.GetMiningTarget(bh.Height, hv.cfg.Miner.AdjustmentInterval, hv.cfg.Miner.MiningInterval)
	if err != nil {
		return fmt.Errorf("error while validating target: %w", err)
	}

	if targetExpected != bh.Target {
		return fmt.Errorf("expected %d target, but header had %d", targetExpected, bh.Target)
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
		return fmt.Errorf("block %x has no coinbase transaction", b.Hash)
	}

	if numCoinbases > 1 {
		return fmt.Errorf("block %x has more than one coinbase transaction", b.Hash)
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

// validateMerkleTree checks that the Merkle tree root hash of the block matches the Merkle root hash in the block header
func (hv *HValidator) validateMerkleTree(b *kernel.Block) error {
	merkletree, err := consensus.NewMerkleTreeFromTxs(b.Transactions, hv.hasher)
	if err != nil {
		return fmt.Errorf("error while constructing merkle tree: %w", err)
	}

	if !bytes.Equal(merkletree.RootHash(), b.Header.MerkleRoot) {
		return fmt.Errorf("block %x has invalid Merkle root", b.Hash)
	}

	return nil
}
