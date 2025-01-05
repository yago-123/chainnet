package storage //nolint:testpackage // don't create separate package for tests

import (
	cerror "github.com/yago-123/chainnet/pkg/error"
	"os"
	"testing"

	"github.com/yago-123/chainnet/pkg/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MockStorageFile = "test-file"

func TestBoltDB_NotFound(t *testing.T) {
	defer os.Remove(MockStorageFile)

	bolt, err := NewBoltDB(MockStorageFile, "block-bucket", "header-bucket", encoding.NewGobEncoder())
	require.NoError(t, err)

	_, err = bolt.GetLastBlock()
	assert.Equal(t, cerror.ErrStorageElementNotFound, err)

	_, err = bolt.GetGenesisBlock()
	assert.Equal(t, cerror.ErrStorageElementNotFound, err)

	_, err = bolt.GetGenesisHeader()
	assert.Equal(t, cerror.ErrStorageElementNotFound, err)

	_, err = bolt.RetrieveBlockByHash([]byte(""))
	assert.Equal(t, cerror.ErrStorageElementNotFound, err)

	_, err = bolt.RetrieveHeaderByHash([]byte(""))
	assert.Equal(t, cerror.ErrStorageElementNotFound, err)
}
