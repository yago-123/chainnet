package mempool

import (
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
)

// MemPoolExplorer is a middleware for MemPool, designed to prevent other packages, such as the network layer,
// from directly accessing its structure. This separation ensures a clear boundary, restricting MemPool manipulation
// to the chain object only
type MemPoolExplorer struct {
	mempool *MemPool
}

func NewMemPoolExplorer(mempool *MemPool) *MemPoolExplorer {
	return &MemPoolExplorer{
		mempool: mempool,
	}
}

func (me *MemPoolExplorer) RetrieveTx(txID string) (*kernel.Transaction, error) {
	me.mempool.mu.Lock()
	defer me.mempool.mu.Unlock()

	if _, ok := me.mempool.txIDs[txID]; !ok {
		return nil, fmt.Errorf("transaction with ID %s not found", txID)
	}

	return me.mempool.txIDs[txID], nil
}
