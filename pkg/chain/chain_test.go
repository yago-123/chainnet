package blockchain //nolint:testpackage // don't create separate package for tests
import (
	"os"
	"testing"

	cerror "github.com/yago-123/chainnet/pkg/error"

	"github.com/yago-123/chainnet/pkg/utxoset"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/storage"
	"github.com/yago-123/chainnet/tests/mocks/consensus"
	mockHash "github.com/yago-123/chainnet/tests/mocks/crypto/hash"
	mockStorage "github.com/yago-123/chainnet/tests/mocks/storage"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var block1 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Header: &kernel.BlockHeader{
		Version:       []byte("1"),
		PrevBlockHash: []byte{},
		MerkleRoot:    []byte{},
		Height:        0,
		Timestamp:     0,
		Target:        1,
		Nonce:         0,
	},
	Transactions: []*kernel.Transaction{
		kernel.NewCoinbaseTransaction("pubkey", 50, 0),
	},
	Hash: []byte("block-1-hash"),
}

var block2 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Header: &kernel.BlockHeader{
		Version:       []byte("1"),
		PrevBlockHash: []byte("block-1-hash"),
		MerkleRoot:    []byte{},
		Height:        1,
		Timestamp:     0,
		Target:        1,
		Nonce:         0,
	},
	Transactions: []*kernel.Transaction{
		kernel.NewCoinbaseTransaction("pubkey", 50, 0),
	},
	Hash: []byte("block-2-hash"),
}

var block3 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Header: &kernel.BlockHeader{
		Version:       []byte("1"),
		PrevBlockHash: []byte("block-2-hash"),
		MerkleRoot:    []byte{},
		Height:        2,
		Timestamp:     0,
		Target:        1,
		Nonce:         0,
	},
	Transactions: []*kernel.Transaction{
		kernel.NewCoinbaseTransaction("pubkey", 50, 0),
	},
	Hash: []byte("block-3-hash"),
}

var block4 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Header: &kernel.BlockHeader{
		Version:       []byte("1"),
		PrevBlockHash: []byte("block-3-hash"),
		MerkleRoot:    []byte{},
		Height:        3,
		Timestamp:     0,
		Target:        1,
		Nonce:         0,
	},
	Transactions: []*kernel.Transaction{
		kernel.NewCoinbaseTransaction("pubkey", 50, 0),
	},
	Hash: []byte("block-4-hash"),
}

// tests the NewBlockchain method when there is not any previous chain addition
func TestBlockchain_InitializationFromScratch(t *testing.T) {
	store := &mockStorage.MockStorage{}
	store.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, cerror.ErrStorageElementNotFound)

	cfg := &config.Config{Logger: logrus.New()}

	chain, err := NewBlockchain(
		cfg,
		store,
		mempool.NewMemPool(1000),
		utxoset.NewUTXOSet(cfg),
		&mockHash.FakeHashing{},
		&consensus.MockHeavyValidator{},
		observer.NewChainSubject(),
		encoding.NewGobEncoder(),
	)

	require.NoError(t, err)
	assert.Equal(t, uint(0), chain.lastHeight)
	assert.Empty(t, chain.lastBlockHash, 0)
	assert.Empty(t, chain.headers, 0)
}

// tests the NewBlockchain method when there has been additions to the chain before
func TestBlockchain_InitializationRecovery(t *testing.T) {
	boltdb, err := storage.NewBoltDB("temp-file", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	require.NoError(t, err)
	defer os.Remove("temp-file")

	// persist headers in store
	require.NoError(t, boltdb.PersistHeader(block1.Hash, *block1.Header))
	require.NoError(t, boltdb.PersistHeader(block2.Hash, *block2.Header))
	require.NoError(t, boltdb.PersistHeader(block3.Hash, *block3.Header))
	require.NoError(t, boltdb.PersistHeader(block4.Hash, *block4.Header))

	// persist blocks in store
	require.NoError(t, boltdb.PersistBlock(*block1))
	require.NoError(t, boltdb.PersistBlock(*block2))
	require.NoError(t, boltdb.PersistBlock(*block3))
	require.NoError(t, boltdb.PersistBlock(*block4))

	mockHashing := &mockHash.MockHashing{}
	mockHashing.
		On("Hash", block4.Header.Assemble()).
		Return([]byte("block-4-hash"), nil)

	cfg := &config.Config{Logger: logrus.New()}
	// initialize chain and make sure that the values are retrieved correctly
	chain, err := NewBlockchain(
		cfg,
		boltdb,
		mempool.NewMemPool(1000),
		utxoset.NewUTXOSet(cfg),
		mockHashing,
		&consensus.MockHeavyValidator{},
		observer.NewChainSubject(),
		encoding.NewGobEncoder(),
	)

	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, []byte("block-4-hash"), chain.lastBlockHash)
	assert.Equal(t, uint(4), chain.lastHeight)
	assert.Len(t, chain.headers, 4)
	assert.Equal(t, []byte("block-3-hash"), chain.headers["block-4-hash"].PrevBlockHash)
}
