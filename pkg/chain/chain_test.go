package blockchain //nolint:testpackage // don't create separate package for tests
import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/kernel"
	"chainnet/tests/mocks/consensus"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockStorage "chainnet/tests/mocks/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var block1 = &kernel.Block{
	Header:       &kernel.BlockHeader{},
	Transactions: []*kernel.Transaction{},
	Hash:         []byte{},
}

var block2 = &kernel.Block{
	Header:       &kernel.BlockHeader{},
	Transactions: []*kernel.Transaction{},
	Hash:         []byte{},
}

var block3 = &kernel.Block{
	Header:       &kernel.BlockHeader{},
	Transactions: []*kernel.Transaction{},
	Hash:         []byte{},
}

var block4 = &kernel.Block{
	Header:       &kernel.BlockHeader{},
	Transactions: []*kernel.Transaction{},
	Hash:         []byte{},
}

// tests the NewBlockchain method when there is not any previous chain addition
func TestBlockchain_InitializationFromScratch(t *testing.T) {
	storage := &mockStorage.MockStorage{}
	storage.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, nil)

	chain, err := NewBlockchain(
		&config.Config{},
		storage,
		&mockHash.MockHashing{},
		&consensus.MockHeavyValidator{},
		observer.NewSubjectObserver(),
	)

	require.NoError(t, err)
	assert.Equal(t, uint(0), chain.lastHeight)
	assert.Len(t, chain.lastBlockHash, 0)
	assert.Len(t, chain.headers, 0)
}

// tests the NewBlockchain method when there has been additions to the chain before
func TestBlockchain_InitializationRecovery(t *testing.T) {
	// write blocks

	// initialize chain

	// make sure that new blocks can be added after

}
