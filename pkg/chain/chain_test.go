package blockchain //nolint:testpackage // don't create separate package for tests
import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"chainnet/tests/mocks/consensus"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockStorage "chainnet/tests/mocks/storage"
	"os"
	"testing"

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
	storage := &mockStorage.MockStorage{}
	storage.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, nil)

	chain, err := NewBlockchain(
		&config.Config{Logger: logrus.New()},
		storage,
		&mockHash.FakeHashing{},
		&consensus.MockHeavyValidator{},
		observer.NewBlockSubject(),
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

	// persist headers in storage
	require.NoError(t, boltdb.PersistHeader(block1.Hash, *block1.Header))
	require.NoError(t, boltdb.PersistHeader(block2.Hash, *block2.Header))
	require.NoError(t, boltdb.PersistHeader(block3.Hash, *block3.Header))
	require.NoError(t, boltdb.PersistHeader(block4.Hash, *block4.Header))

	mockHashing := &mockHash.MockHashing{}
	mockHashing.
		On("Hash", block4.Header.Assemble()).
		Return([]byte("block-4-hash"), nil)

	// initialize chain and make sure that the values are retrieved correctly
	chain, err := NewBlockchain(
		&config.Config{Logger: logrus.New()},
		boltdb,
		mockHashing,
		&consensus.MockHeavyValidator{},
		observer.NewBlockSubject(),
	)

	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, []byte("block-4-hash"), chain.lastBlockHash)
	assert.Equal(t, uint(4), chain.lastHeight)
	assert.Len(t, chain.headers, 4)
	assert.Equal(t, []byte("block-3-hash"), chain.headers["block-4-hash"].PrevBlockHash)
}
