package storage

import (
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	boltdb "github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestBoltDB_GetLastBlock(t *testing.T) {
	type fields struct {
		db       *boltdb.DB
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    *kernel.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bolt := &BoltDB{
				db:       tt.fields.db,
				bucket:   tt.fields.bucket,
				encoding: tt.fields.encoding,
				logger:   tt.fields.logger,
			}
			got, err := bolt.GetLastBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoltDB_NumberOfBlocks(t *testing.T) {
	type fields struct {
		db       *boltdb.DB
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    uint
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bolt := &BoltDB{
				db:       tt.fields.db,
				bucket:   tt.fields.bucket,
				encoding: tt.fields.encoding,
				logger:   tt.fields.logger,
			}
			got, err := bolt.NumberOfBlocks()
			if (err != nil) != tt.wantErr {
				t.Errorf("NumberOfBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NumberOfBlocks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoltDB_PersistBlock(t *testing.T) {
	type fields struct {
		db       *boltdb.DB
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	type args struct {
		block kernel.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bolt := &BoltDB{
				db:       tt.fields.db,
				bucket:   tt.fields.bucket,
				encoding: tt.fields.encoding,
				logger:   tt.fields.logger,
			}
			if err := bolt.PersistBlock(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("PersistBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoltDB_RetrieveBlockByHash(t *testing.T) {
	type fields struct {
		db       *boltdb.DB
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	type args struct {
		hash []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *kernel.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bolt := &BoltDB{
				db:       tt.fields.db,
				bucket:   tt.fields.bucket,
				encoding: tt.fields.encoding,
				logger:   tt.fields.logger,
			}
			got, err := bolt.RetrieveBlockByHash(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveBlockByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveBlockByHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoltDB_bucketExists(t *testing.T) {
	type fields struct {
		db       *boltdb.DB
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	type args struct {
		bucketName string
		tx         *boltdb.Tx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
		want1  *boltdb.Bucket
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bolt := &BoltDB{
				db:       tt.fields.db,
				bucket:   tt.fields.bucket,
				encoding: tt.fields.encoding,
				logger:   tt.fields.logger,
			}
			got, got1 := bolt.bucketExists(tt.args.bucketName, tt.args.tx)
			if got != tt.want {
				t.Errorf("bucketExists() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("bucketExists() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewBoltDB(t *testing.T) {
	type args struct {
		dbFile   string
		bucket   string
		encoding encoding.Encoding
		logger   *logrus.Logger
	}
	tests := []struct {
		name string
		args args
		want *BoltDB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewBoltDB(tt.args.dbFile, tt.args.bucket, tt.args.encoding, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBoltDB() = %v, want %v", got, tt.want)
			}
		})
	}
}
