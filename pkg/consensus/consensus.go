package consensus

import (
	"github.com/yago-123/chainnet/pkg/kernel"
)

// LightValidator represents a validator that does not require having the whole chain downloaded locally
// like for example the ones performed by wallets before sending transactions to the nodes and miners
type LightValidator interface {
	ValidateTxLight(tx *kernel.Transaction) error
	ValidateHeader(bh *kernel.BlockHeader) error
}

// HeavyValidator performs the same validations as LightValidator but also validates the previous
// information like the validity of the chain and transactions without funds. This validator is used
// by nodes and miners
type HeavyValidator interface {
	ValidateTx(tx *kernel.Transaction) error
	ValidateHeader(bh *kernel.BlockHeader) error
	ValidateBlock(b *kernel.Block) error
}
