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

	LastBlockKey  = "lastblock"
	FirstBlockKey = "firstblock"
)

type BoltDB struct {
	db     *boltdb.DB
	bucket string

	encoding encoding.Encoding
}

func NewBoltDB(dbFile string, bucket string, encoding encoding.Encoding) (*BoltDB, error) {
	db, err := boltdb.Open(dbFile, BoltDBCreationMode, &boltdb.Options{Timeout: BoltDBTimeout})
	if err != nil {
		return nil, fmt.Errorf("error opening bolt storage: %w", err)
	}
	return &BoltDB{db, bucket, encoding}, nil
}

func (bolt *BoltDB) NumberOfBlocks() (uint, error) {
	var numKeys uint

	err := bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if exists {
			numKeys = uint(bucket.Stats().KeyN)
		}
		return nil
	})

	return numKeys, err
}

func (bolt *BoltDB) PersistBlock(block kernel.Block) error {
	var err error

	dataBlock, err := bolt.encoding.SerializeBlock(block)
	if err != nil {
		return fmt.Errorf("error serializing block %s: %w", string(block.Hash), err)
	}

	err = bolt.db.Update(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		// create bucket if does not exist yet
		if !exists {
			bucket, err = tx.CreateBucket([]byte(bolt.bucket))
			if err != nil {
				return fmt.Errorf("error creating bucket: %w", err)
			}

			// if the bucket did not exist, this is the genesis block
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

func (bolt *BoltDB) GetLastBlock() (*kernel.Block, error) {
	var err error
	var lastBlock []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if !exists {
			return fmt.Errorf("error getting last block: bucket %s does not exist", bolt.bucket)
		}

		lastBlock = bucket.Get([]byte(LastBlockKey))

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	return bolt.encoding.DeserializeBlock(lastBlock)
}

func (bolt *BoltDB) GetGenesisBlock() (*kernel.Block, error) {
	var genesisBlock []byte

	err := bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if !exists {
			return fmt.Errorf("error getting genesis block: bucket %s does not exist", bolt.bucket)
		}

		genesisBlock = bucket.Get([]byte(FirstBlockKey))

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	return bolt.encoding.DeserializeBlock(genesisBlock)
}

func (bolt *BoltDB) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	var err error
	var blockBytes []byte
	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if !exists {
			return fmt.Errorf("error retrieving block %s: bucket %s does not exist", string(hash), bolt.bucket)
		}

		blockBytes = bucket.Get(hash)

		return nil
	})

	if err != nil {
		return &kernel.Block{}, err
	}

	return bolt.encoding.DeserializeBlock(blockBytes)
}

func (bolt *BoltDB) bucketExists(bucketName string, tx *boltdb.Tx) (bool, *boltdb.Bucket) {
	bucket := tx.Bucket([]byte(bucketName))
	return bucket != nil, bucket
}
