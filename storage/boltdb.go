package storage

import (
	"chainnet/block"
	"chainnet/encoding"
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
		bucket := tx.Bucket([]byte(bolt.bucket))
		if bucket != nil {
			numKeys = uint(bucket.Stats().KeyN)
		}
		return nil
	})

	return numKeys, err
}

func (bolt *BoltDB) PersistBlock(block block.Block) error {
	var bucket *boltdb.Bucket
	var err error

	dataBlock, err := bolt.encoding.SerializeBlock(block)
	if err != nil {
		return err
	}

	err = bolt.db.Update(func(tx *boltdb.Tx) error {
		bucket = tx.Bucket([]byte(bolt.bucket))
		// create bucket if does not exist yet
		if bucket == nil {
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
