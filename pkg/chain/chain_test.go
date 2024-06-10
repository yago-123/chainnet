package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"chainnet/pkg/consensus"
	"chainnet/pkg/storage"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestBlockchain_AddBlock(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	type args struct {
		transactions []*block.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *block.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				Chain:         tt.fields.Chain,
				lastBlockHash: tt.fields.lastBlockHash,
				consensus:     tt.fields.consensus,
				storage:       tt.fields.storage,
				logger:        tt.fields.logger,
				cfg:           tt.fields.cfg,
			}
			got, err := bc.AddBlock(tt.args.transactions)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchain_CreateIterator(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   Iterator
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				Chain:         tt.fields.Chain,
				lastBlockHash: tt.fields.lastBlockHash,
				consensus:     tt.fields.consensus,
				storage:       tt.fields.storage,
				logger:        tt.fields.logger,
				cfg:           tt.fields.cfg,
			}
			if got := bc.CreateIterator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateIterator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchain_FindUTXO(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	type args struct {
		address string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []block.TxOutput
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				Chain:         tt.fields.Chain,
				lastBlockHash: tt.fields.lastBlockHash,
				consensus:     tt.fields.consensus,
				storage:       tt.fields.storage,
				logger:        tt.fields.logger,
				cfg:           tt.fields.cfg,
			}
			if got := bc.FindUTXO(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUTXO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchain_FindUnspentTransactions(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	type args struct {
		address string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*block.Transaction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				Chain:         tt.fields.Chain,
				lastBlockHash: tt.fields.lastBlockHash,
				consensus:     tt.fields.consensus,
				storage:       tt.fields.storage,
				logger:        tt.fields.logger,
				cfg:           tt.fields.cfg,
			}
			if got := bc.FindUnspentTransactions(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUnspentTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchain_GetBlock(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *block.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				Chain:         tt.fields.Chain,
				lastBlockHash: tt.fields.lastBlockHash,
				consensus:     tt.fields.consensus,
				storage:       tt.fields.storage,
				logger:        tt.fields.logger,
				cfg:           tt.fields.cfg,
			}
			got, err := bc.GetBlock(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIteratorStruct_GetNextBlock(t *testing.T) {
	type fields struct {
		prevBlockHash []byte
		storage       storage.Storage
		cfg           *config.Config
	}
	tests := []struct {
		name    string
		fields  fields
		want    *block.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := &IteratorStruct{
				prevBlockHash: tt.fields.prevBlockHash,
				storage:       tt.fields.storage,
				cfg:           tt.fields.cfg,
			}
			got, err := it.GetNextBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNextBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIteratorStruct_HasNext(t *testing.T) {
	type fields struct {
		prevBlockHash []byte
		storage       storage.Storage
		cfg           *config.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := &IteratorStruct{
				prevBlockHash: tt.fields.prevBlockHash,
				storage:       tt.fields.storage,
				cfg:           tt.fields.cfg,
			}
			if got := it.HasNext(); got != tt.want {
				t.Errorf("HasNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlockchain(t *testing.T) {
	type args struct {
		cfg         *config.Config
		consensus   consensus.Consensus
		persistence storage.Storage
	}
	tests := []struct {
		name string
		args args
		want *Blockchain
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchain(tt.args.cfg, tt.args.consensus, tt.args.persistence); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIterator(t *testing.T) {
	type args struct {
		lastBlockHash []byte
		storage       storage.Storage
		cfg           *config.Config
	}
	tests := []struct {
		name string
		args args
		want *IteratorStruct
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIterator(tt.args.lastBlockHash, tt.args.storage, tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIterator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOutputSpent(t *testing.T) {
	type args struct {
		spentTXOs map[string][]int
		txID      string
		outIdx    int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOutputSpent(tt.args.spentTXOs, tt.args.txID, tt.args.outIdx); got != tt.want {
				t.Errorf("isOutputSpent() = %v, want %v", got, tt.want)
			}
		})
	}
}
