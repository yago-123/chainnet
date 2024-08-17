package storage

import (
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"fmt"
	"time"

	boltdb "github.com/boltdb/bolt"
)

const (
	BoltDBCreationMode = 0600
	BoltDBTimeout      = 5 * time.Second
)

type BoltDB struct {
	db           *boltdb.DB
	blockBucket  string
	headerBucket string

	encoding encoding.Encoding
}

func NewBoltDB(dbFile, blockBucket, headerBucket string, encoding encoding.Encoding) (*BoltDB, error) {
	db, err := boltdb.Open(dbFile, BoltDBCreationMode, &boltdb.Options{Timeout: BoltDBTimeout})
	if err != nil {
		return nil, fmt.Errorf("error opening bolt storage: %w", err)
	}

	return &BoltDB{
		db:           db,
		blockBucket:  blockBucket,
		headerBucket: headerBucket,
		encoding:     encoding,
	}, nil
}

func (bolt *BoltDB) PersistBlock(block kernel.Block) error {
	var err error

	dataBlock, err := bolt.encoding.SerializeBlock(block)
	if err != nil {
		return fmt.Errorf("error serializing block %s: %w", string(block.Hash), err)
	}

	err = bolt.db.Update(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.blockBucket, tx)
		// create blockBucket if does not exist yet
		if !exists {
			bucket, err = tx.CreateBucket([]byte(bolt.blockBucket))
			if err != nil {
				return fmt.Errorf("error creating blockBucket: %w", err)
			}

			// if the blockBucket did not exist, this is the genesis block
			// todo() handle this part when p2p and state restoration is tackled
			err = bucket.Put([]byte(FirstBlockKey), dataBlock)
			if err != nil {
				return fmt.Errorf("error writing first block %s: %w", string(block.Hash), err)
			}
		}

		// add the new k/v
		err = bucket.Put(block.Hash, dataBlock)
		if err != nil {
			return fmt.Errorf("error writing block %s: %w", string(block.Hash), err)
		}

		// update key pointing to last block
		err = bucket.Put([]byte(LastBlockKey), dataBlock)
		if err != nil {
			return fmt.Errorf("error writing last block %s: %w", string(block.Hash), err)
		}

		return nil
	})

	return err
}

func (bolt *BoltDB) PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error {
	var err error

	dataHeader, err := bolt.encoding.SerializeHeader(blockHeader)
	if err != nil {
		return fmt.Errorf("error serializing block header: %w", err)
	}

	err = bolt.db.Update(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.headerBucket, tx)
		// create headerBucket if does not exist yet
		if !exists {
			bucket, err = tx.CreateBucket([]byte(bolt.headerBucket))
			if err != nil {
				return fmt.Errorf("error creating header bucket: %w", err)
			}

			// if the headerBucket does not exist, this is the genesis block header
			err = bucket.Put([]byte(FirstHeaderKey), dataHeader)
			if err != nil {
				return fmt.Errorf("error writing first header %s: %w", string(blockHash), err)
			}
		}

		// add the new k/v
		err = bucket.Put(blockHash, dataHeader)
		if err != nil {
			return fmt.Errorf("error writing block %s: %w", string(blockHash), err)
		}

		// update key pointing to last header
		err = bucket.Put([]byte(LastHeaderKey), dataHeader)
		if err != nil {
			return fmt.Errorf("error writing last header %s: %w", string(blockHash), err)
		}

		// update key pointing to last block hash, this value is updated when persisting headers only given that
		// the first persisted field when adding a block to the chain is the header
		err = bucket.Put([]byte(LastBlockHashKey), blockHash)
		if err != nil {
			return fmt.Errorf("error writing last block hash %s: %w", string(blockHash), err)
		}

		return nil
	})
	return err
}

func (bolt *BoltDB) GetLastBlock() (*kernel.Block, error) {
	var err error
	var lastBlock []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.blockBucket, tx)
		if !exists {
			return ErrNotFound
		}

		lastBlock = bucket.Get([]byte(LastBlockKey))

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	if len(lastBlock) == 0 {
		return &kernel.Block{}, ErrNotFound
	}
	return bolt.encoding.DeserializeBlock(lastBlock)
}

func (bolt *BoltDB) GetLastHeader() (*kernel.BlockHeader, error) {
	var err error
	var lastHeader []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.headerBucket, tx)
		if !exists {
			return ErrNotFound
		}

		// if the bucket exists, at least one header has been written
		if exists {
			lastHeader = bucket.Get([]byte(LastHeaderKey))
		}

		// if does not exist, genesis block have not been created yet

		return nil
	})

	if err != nil {
		return &kernel.BlockHeader{}, err
	}

	if len(lastHeader) == 0 {
		return &kernel.BlockHeader{}, ErrNotFound
	}
	return bolt.encoding.DeserializeHeader(lastHeader)
}

func (bolt *BoltDB) GetLastBlockHash() ([]byte, error) {
	var err error
	var lastBlockHash []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.headerBucket, tx)
		// if the last block hash does not exists yet, the genesis block has not been created yet
		if !exists {
			return ErrNotFound
		}

		if exists {
			lastBlockHash = bucket.Get([]byte(LastBlockHashKey))
		}

		return nil
	})

	if err != nil {
		return []byte{}, err
	}

	if len(lastBlockHash) == 0 {
		return []byte{}, ErrNotFound
	}
	return lastBlockHash, nil
}

func (bolt *BoltDB) GetGenesisBlock() (*kernel.Block, error) {
	var genesisBlock []byte

	err := bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.blockBucket, tx)
		if !exists {
			return ErrNotFound
		}

		genesisBlock = bucket.Get([]byte(FirstBlockKey))

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	if len(genesisBlock) == 0 {
		return &kernel.Block{}, ErrNotFound
	}
	return bolt.encoding.DeserializeBlock(genesisBlock)
}

func (bolt *BoltDB) GetGenesisHeader() (*kernel.BlockHeader, error) {
	var genesisHeader []byte

	err := bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.headerBucket, tx)
		if !exists {
			return ErrNotFound
		}

		genesisHeader = bucket.Get([]byte(FirstHeaderKey))

		return nil
	})

	if err != nil {
		return &kernel.BlockHeader{}, err
	}

	if len(genesisHeader) == 0 {
		return &kernel.BlockHeader{}, ErrNotFound
	}
	return bolt.encoding.DeserializeHeader(genesisHeader)
}

func (bolt *BoltDB) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	var err error
	var blockBytes []byte
	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.blockBucket, tx)
		if !exists {
			return ErrNotFound
		}

		blockBytes = bucket.Get(hash)

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	if len(blockBytes) == 0 {
		return &kernel.Block{}, ErrNotFound
	}
	return bolt.encoding.DeserializeBlock(blockBytes)
}

func (bolt *BoltDB) RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error) {
	var err error
	var headerBytes []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bucketExists(bolt.headerBucket, tx)
		if !exists {
			return ErrNotFound
		}

		headerBytes = bucket.Get(hash)

		return nil
	})

	if err != nil {
		return &kernel.BlockHeader{}, err
	}

	if len(headerBytes) == 0 {
		return &kernel.BlockHeader{}, ErrNotFound
	}
	return bolt.encoding.DeserializeHeader(headerBytes)
}

func (bolt *BoltDB) BlockObserverID() string {
	return StorageObserverID
}

func (bolt *BoltDB) OnBlockAddition(block *kernel.Block) {
	// async function because we don't want to create a deadlock (observer is sync by default)
	// todo(): had to remove the async because apparently during synchronization with p2p some cases were failing (recheck)
	// go func() {
	err := bolt.PersistBlock(*block)
	if err != nil {
		// todo(): add logging about the issue, if this fails, will eventually be pulled and stored again by P2P
		return
	}
	// }()
}

func (bolt *BoltDB) Close() error {
	return bolt.db.Close()
}

func bucketExists(bucketName string, tx *boltdb.Tx) (bool, *boltdb.Bucket) {
	bucket := tx.Bucket([]byte(bucketName))
	return bucket != nil, bucket
}
