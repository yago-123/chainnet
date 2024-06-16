package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"chainnet/pkg/consensus"
	"chainnet/pkg/storage"
	mockConsensus "chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	mockUtil "chainnet/tests/mocks/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestBlockchain_AddBlockWithoutErrors(t *testing.T) {
	bc := NewBlockchain(
		config.NewConfig(logrus.New(), 1, 1, ""),
		&mockConsensus.MockConsensus{},
		&mockStorage.MockStorage{},
	)

	coinbaseTx := []*block.Transaction{
		{
			ID: nil,
			Vin: []block.TxInput{
				{
					Txid:      []byte{},
					Vout:      -1,
					ScriptSig: "randomSig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       block.COINBASE_AMOUNT,
					ScriptPubKey: "pubKey",
				},
			},
		},
	}

	secondTx := []*block.Transaction{
		{
			ID: []byte("second-tx-id"),
			Vin: []block.TxInput{
				{
					Txid:      []byte("random"),
					Vout:      0,
					ScriptSig: "random-script-sig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "random-pub-key",
				},
				{
					Amount:       200,
					ScriptPubKey: "random-pub-key-2",
				},
			},
		},
	}

	thirdTx := []*block.Transaction{
		{
			ID: []byte("third-tx-id"),
			Vin: []block.TxInput{
				{
					Txid:      []byte("random"),
					Vout:      0,
					ScriptSig: "random-script-sig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       101,
					ScriptPubKey: "random-pub-key-3",
				},
				{
					Amount:       201,
					ScriptPubKey: "random-pub-key-4",
				},
			},
		},
	}

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(0), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte{})).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte{})).
		Return([]byte("genesis-block-hash"), uint(1), nil)

	// add genesis block
	blockAdded, err := bc.AddBlock(coinbaseTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, 0, len(blockAdded.PrevBlockHash), "genesis blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("genesis-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("genesis-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 1, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "genesis-block-hash", bc.Chain[0], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(1), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("genesis-block-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("genesis-block-hash"))).
		Return([]byte("second-block-hash"), uint(1), nil)

	// add another block
	blockAdded, err = bc.AddBlock(secondTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("genesis-block-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("second-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("second-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 2, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "second-block-hash", bc.Chain[1], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(2), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("second-block-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("second-block-hash"))).
		Return([]byte("third-block-hash"), uint(1), nil)

	// add another block
	blockAdded, err = bc.AddBlock(thirdTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("second-block-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("third-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("third-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 3, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "third-block-hash", bc.Chain[2], "blockchain chain not updated with new blockAdded hash")
}

func TestBlockchain_AddBlockWithErrors(t *testing.T) {

}

func TestBlockchain_AddBlockWithInvalidTransaction(t *testing.T) {

}

func TestBlockchain_CalculateAddressBalance(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		want    int
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
			got, err := bc.CalculateAddressBalance(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateAddressBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateAddressBalance() got = %v, want %v", got, tt.want)
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

func TestBlockchain_FindAmountSpendableOutputs(t *testing.T) {
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
		amount  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		want1   map[string][]int
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
			got, got1, err := bc.FindAmountSpendableOutputs(tt.args.address, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAmountSpendableOutputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindAmountSpendableOutputs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("FindAmountSpendableOutputs() got1 = %v, want %v", got1, tt.want1)
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
		name    string
		fields  fields
		args    args
		want    []block.TxOutput
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
			got, err := bc.FindUTXO(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindUTXO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUTXO() got = %v, want %v", got, tt.want)
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
		name    string
		fields  fields
		args    args
		want    []*block.Transaction
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
			got, err := bc.FindUnspentTransactions(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindUnspentTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUnspentTransactions() got = %v, want %v", got, tt.want)
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

func TestBlockchain_MineBlock(t *testing.T) {
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
		name   string
		fields fields
		args   args
		want   *block.Block
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
			if got := bc.MineBlock(tt.args.transactions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MineBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchain_NewUTXOTransaction(t *testing.T) {
	type fields struct {
		Chain         []string
		lastBlockHash []byte
		consensus     consensus.Consensus
		storage       storage.Storage
		logger        *logrus.Logger
		cfg           *config.Config
	}
	type args struct {
		from   string
		to     string
		amount int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *block.Transaction
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
			got, err := bc.NewTransaction(tt.args.from, tt.args.to, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransaction() got = %v, want %v", got, tt.want)
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
