package storage

import (
	"chainnet/block"
	"chainnet/encoding"
	"fmt"
	boltdb "github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"time"
)

const LastBlockKey = "lastblock"

type BoltDB struct {
	db     *boltdb.DB
	bucket string

	encoding encoding.Encoding
	logger   *logrus.Logger
}

func NewBoltDB(dbFile string, bucket string, encoding encoding.Encoding, logger *logrus.Logger) *BoltDB {
	db, err := boltdb.Open(dbFile, 0600, &boltdb.Options{Timeout: 5 * time.Second})
	if err != nil {
		logger.Errorf("Error opening BoltDB: %s", err)
	}

	return &BoltDB{db, bucket, encoding, logger}
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

func (bolt *BoltDB) PersistBlock(block block.Block) error {
	var err error

	dataBlock, err := bolt.encoding.SerializeBlock(block)
	if err != nil {
		return err
	}

	err = bolt.db.Update(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		// create bucket if does not exist yet
		if !exists {
			bucket, err = tx.CreateBucket([]byte(bolt.bucket))
			if err != nil {
				return err
			}
		}

		// add the new k/v
		bucket.Put(block.Hash, dataBlock)
		// update key pointing to last block
		bucket.Put([]byte(LastBlockKey), block.Hash)

		return nil
	})

	return err
}

func (bolt *BoltDB) GetLastBlock() (*block.Block, error) {
	var err error
	var lastBlock []byte

	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if !exists {
			return fmt.Errorf("bucket does not exist")
		}

		lastBlock = bucket.Get([]byte(LastBlockKey))

		return nil
	})

	if err != nil {
		return &block.Block{}, err
	}

	return bolt.encoding.DeserializeBlock(lastBlock)
}

func (bolt *BoltDB) RetrieveBlockByHash(hash []byte) (*block.Block, error) {
	var err error
	var blockBytes []byte
	err = bolt.db.View(func(tx *boltdb.Tx) error {
		exists, bucket := bolt.bucketExists(bolt.bucket, tx)
		if !exists {
			return fmt.Errorf("bucket does not exist")
		}

		blockBytes = bucket.Get(hash)

		return nil
	})

	if err != nil {
		return &block.Block{}, err
	}

	return bolt.encoding.DeserializeBlock(blockBytes)
}

func (bolt *BoltDB) bucketExists(bucketName string, tx *boltdb.Tx) (bool, *boltdb.Bucket) {
	bucket := tx.Bucket([]byte(bucketName))
	return bucket != nil, bucket
}
