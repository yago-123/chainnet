package encoding

import (
	. "chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"reflect"
	"testing"
)

var Block1 = Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-1-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-2"),
			},
		},
	},
	PrevBlockHash: []byte("genesish-block-hash"),
	Nonce:         1,
	Hash:          []byte("block-hash-1"),
}

var Block1Encoded = []byte{98, 127, 3, 1, 1, 5, 66, 108, 111, 99, 107, 1, 255, 128, 0, 1, 6, 1, 9, 84, 105, 109, 101, 115, 116, 97, 109, 112, 1, 4, 0, 1, 12, 84, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 115, 1, 255, 140, 0, 1, 13, 80, 114, 101, 118, 66, 108, 111, 99, 107, 72, 97, 115, 104, 1, 10, 0, 1, 6, 84, 97, 114, 103, 101, 116, 1, 6, 0, 1, 5, 78, 111, 110, 99, 101, 1, 6, 0, 1, 4, 72, 97, 115, 104, 1, 10, 0, 0, 0, 36, 255, 139, 2, 1, 1, 21, 91, 93, 42, 107, 101, 114, 110, 101, 108, 46, 84, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 1, 255, 140, 0, 1, 255, 130, 0, 0, 38, 255, 129, 3, 1, 2, 255, 130, 0, 1, 3, 1, 2, 73, 68, 1, 10, 0, 1, 3, 86, 105, 110, 1, 255, 134, 0, 1, 4, 86, 111, 117, 116, 1, 255, 138, 0, 0, 0, 31, 255, 133, 2, 1, 1, 16, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 73, 110, 112, 117, 116, 1, 255, 134, 0, 1, 255, 132, 0, 0, 64, 255, 131, 3, 1, 1, 7, 84, 120, 73, 110, 112, 117, 116, 1, 255, 132, 0, 1, 4, 1, 4, 84, 120, 105, 100, 1, 10, 0, 1, 4, 86, 111, 117, 116, 1, 6, 0, 1, 9, 83, 99, 114, 105, 112, 116, 83, 105, 103, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 32, 255, 137, 2, 1, 1, 17, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 138, 0, 1, 255, 136, 0, 0, 61, 255, 135, 3, 1, 1, 8, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 136, 0, 1, 3, 1, 6, 65, 109, 111, 117, 110, 116, 1, 6, 0, 1, 12, 83, 99, 114, 105, 112, 116, 80, 117, 98, 75, 101, 121, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 116, 255, 128, 2, 1, 1, 31, 99, 111, 105, 110, 98, 97, 115, 101, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 49, 45, 105, 100, 1, 1, 0, 1, 1, 1, 50, 1, 20, 112, 117, 98, 75, 101, 121, 45, 50, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 0, 1, 19, 103, 101, 110, 101, 115, 105, 115, 104, 45, 98, 108, 111, 99, 107, 45, 104, 97, 115, 104, 2, 1, 1, 12, 98, 108, 111, 99, 107, 45, 104, 97, 115, 104, 45, 49, 0}

var Block2 = Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-2-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-3"),
			},
		},
		{
			ID: []byte("regular-transaction-block-2-id"),
			Vin: []TxInput{
				NewInput([]byte("coinbase-transaction-block-1-id"), 0, "pubKey-2", "pubKey-2"),
			},
			Vout: []TxOutput{
				NewOutput(2, script.P2PK, "pubKey-3"),
				NewOutput(3, script.P2PK, "pubKey-4"),
				NewOutput(44, script.P2PK, "pubKey-5"),
				NewOutput(1, script.P2PK, "pubKey-2"),
			},
		},
	},
	PrevBlockHash: []byte("block-2-hash"),
	Nonce:         1,
	Hash:          []byte("block-hash-2"),
}

var Block2Encoded = []byte{98, 127, 3, 1, 1, 5, 66, 108, 111, 99, 107, 1, 255, 128, 0, 1, 6, 1, 9, 84, 105, 109, 101, 115, 116, 97, 109, 112, 1, 4, 0, 1, 12, 84, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 115, 1, 255, 140, 0, 1, 13, 80, 114, 101, 118, 66, 108, 111, 99, 107, 72, 97, 115, 104, 1, 10, 0, 1, 6, 84, 97, 114, 103, 101, 116, 1, 6, 0, 1, 5, 78, 111, 110, 99, 101, 1, 6, 0, 1, 4, 72, 97, 115, 104, 1, 10, 0, 0, 0, 36, 255, 139, 2, 1, 1, 21, 91, 93, 42, 107, 101, 114, 110, 101, 108, 46, 84, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 1, 255, 140, 0, 1, 255, 130, 0, 0, 38, 255, 129, 3, 1, 2, 255, 130, 0, 1, 3, 1, 2, 73, 68, 1, 10, 0, 1, 3, 86, 105, 110, 1, 255, 134, 0, 1, 4, 86, 111, 117, 116, 1, 255, 138, 0, 0, 0, 31, 255, 133, 2, 1, 1, 16, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 73, 110, 112, 117, 116, 1, 255, 134, 0, 1, 255, 132, 0, 0, 64, 255, 131, 3, 1, 1, 7, 84, 120, 73, 110, 112, 117, 116, 1, 255, 132, 0, 1, 4, 1, 4, 84, 120, 105, 100, 1, 10, 0, 1, 4, 86, 111, 117, 116, 1, 6, 0, 1, 9, 83, 99, 114, 105, 112, 116, 83, 105, 103, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 32, 255, 137, 2, 1, 1, 17, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 138, 0, 1, 255, 136, 0, 0, 61, 255, 135, 3, 1, 1, 8, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 136, 0, 1, 3, 1, 6, 65, 109, 111, 117, 110, 116, 1, 6, 0, 1, 12, 83, 99, 114, 105, 112, 116, 80, 117, 98, 75, 101, 121, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 254, 1, 84, 255, 128, 2, 2, 1, 31, 99, 111, 105, 110, 98, 97, 115, 101, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 50, 45, 105, 100, 1, 1, 0, 1, 1, 1, 50, 1, 20, 112, 117, 98, 75, 101, 121, 45, 51, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 51, 0, 0, 1, 30, 114, 101, 103, 117, 108, 97, 114, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 50, 45, 105, 100, 1, 1, 1, 31, 99, 111, 105, 110, 98, 97, 115, 101, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 49, 45, 105, 100, 2, 8, 112, 117, 98, 75, 101, 121, 45, 50, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 1, 4, 1, 2, 1, 20, 112, 117, 98, 75, 101, 121, 45, 51, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 51, 0, 1, 3, 1, 20, 112, 117, 98, 75, 101, 121, 45, 52, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 52, 0, 1, 44, 1, 20, 112, 117, 98, 75, 101, 121, 45, 53, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 53, 0, 1, 1, 1, 20, 112, 117, 98, 75, 101, 121, 45, 50, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 0, 1, 12, 98, 108, 111, 99, 107, 45, 50, 45, 104, 97, 115, 104, 2, 1, 1, 12, 98, 108, 111, 99, 107, 45, 104, 97, 115, 104, 45, 50, 0}

var Transaction1 = Transaction{
	ID: []byte("coinbase-transaction-block-1-id"),
	Vin: []TxInput{
		NewCoinbaseInput(),
	},
	Vout: []TxOutput{
		NewOutput(CoinbaseReward, script.P2PK, "pubKey-2"),
	},
}

var Transaction1Encoded = []byte{38, 255, 129, 3, 1, 2, 255, 130, 0, 1, 3, 1, 2, 73, 68, 1, 10, 0, 1, 3, 86, 105, 110, 1, 255, 134, 0, 1, 4, 86, 111, 117, 116, 1, 255, 138, 0, 0, 0, 31, 255, 133, 2, 1, 1, 16, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 73, 110, 112, 117, 116, 1, 255, 134, 0, 1, 255, 132, 0, 0, 64, 255, 131, 3, 1, 1, 7, 84, 120, 73, 110, 112, 117, 116, 1, 255, 132, 0, 1, 4, 1, 4, 84, 120, 105, 100, 1, 10, 0, 1, 4, 86, 111, 117, 116, 1, 6, 0, 1, 9, 83, 99, 114, 105, 112, 116, 83, 105, 103, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 32, 255, 137, 2, 1, 1, 17, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 138, 0, 1, 255, 136, 0, 0, 61, 255, 135, 3, 1, 1, 8, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 136, 0, 1, 3, 1, 6, 65, 109, 111, 117, 110, 116, 1, 6, 0, 1, 12, 83, 99, 114, 105, 112, 116, 80, 117, 98, 75, 101, 121, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 76, 255, 130, 1, 31, 99, 111, 105, 110, 98, 97, 115, 101, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 49, 45, 105, 100, 1, 1, 0, 1, 1, 1, 50, 1, 20, 112, 117, 98, 75, 101, 121, 45, 50, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 0}

var Transaction2 = Transaction{
	ID: []byte("regular-transaction-block-2-id"),
	Vin: []TxInput{
		NewInput([]byte("coinbase-transaction-block-1-id"), 0, "pubKey-2", "pubKey-2"),
	},
	Vout: []TxOutput{
		NewOutput(2, script.P2PK, "pubKey-3"),
		NewOutput(3, script.P2PK, "pubKey-4"),
		NewOutput(44, script.P2PK, "pubKey-5"),
		NewOutput(1, script.P2PK, "pubKey-2"),
	},
}

var Transaction2Encoded = []byte{38, 255, 129, 3, 1, 2, 255, 130, 0, 1, 3, 1, 2, 73, 68, 1, 10, 0, 1, 3, 86, 105, 110, 1, 255, 134, 0, 1, 4, 86, 111, 117, 116, 1, 255, 138, 0, 0, 0, 31, 255, 133, 2, 1, 1, 16, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 73, 110, 112, 117, 116, 1, 255, 134, 0, 1, 255, 132, 0, 0, 64, 255, 131, 3, 1, 1, 7, 84, 120, 73, 110, 112, 117, 116, 1, 255, 132, 0, 1, 4, 1, 4, 84, 120, 105, 100, 1, 10, 0, 1, 4, 86, 111, 117, 116, 1, 6, 0, 1, 9, 83, 99, 114, 105, 112, 116, 83, 105, 103, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 32, 255, 137, 2, 1, 1, 17, 91, 93, 107, 101, 114, 110, 101, 108, 46, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 138, 0, 1, 255, 136, 0, 0, 61, 255, 135, 3, 1, 1, 8, 84, 120, 79, 117, 116, 112, 117, 116, 1, 255, 136, 0, 1, 3, 1, 6, 65, 109, 111, 117, 110, 116, 1, 6, 0, 1, 12, 83, 99, 114, 105, 112, 116, 80, 117, 98, 75, 101, 121, 1, 12, 0, 1, 6, 80, 117, 98, 75, 101, 121, 1, 12, 0, 0, 0, 255, 233, 255, 130, 1, 30, 114, 101, 103, 117, 108, 97, 114, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 50, 45, 105, 100, 1, 1, 1, 31, 99, 111, 105, 110, 98, 97, 115, 101, 45, 116, 114, 97, 110, 115, 97, 99, 116, 105, 111, 110, 45, 98, 108, 111, 99, 107, 45, 49, 45, 105, 100, 2, 8, 112, 117, 98, 75, 101, 121, 45, 50, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 1, 4, 1, 2, 1, 20, 112, 117, 98, 75, 101, 121, 45, 51, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 51, 0, 1, 3, 1, 20, 112, 117, 98, 75, 101, 121, 45, 52, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 52, 0, 1, 44, 1, 20, 112, 117, 98, 75, 101, 121, 45, 53, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 53, 0, 1, 1, 1, 20, 112, 117, 98, 75, 101, 121, 45, 50, 32, 79, 80, 95, 67, 72, 69, 67, 75, 83, 73, 71, 1, 8, 112, 117, 98, 75, 101, 121, 45, 50, 0, 0}

func TestGobEncoder_DeserializeBlock(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Block
		wantErr bool
	}{
		{"Deserialize block 1", args{Block1Encoded}, &Block1, false},
		{"Deserialize block 2", args{Block2Encoded}, &Block2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{}
			got, err := gobenc.DeserializeBlock(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGobEncoder_DeserializeTransaction(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Transaction
		wantErr bool
	}{
		{"Deserialize transaction 1", args{Transaction1Encoded}, &Transaction1, false},
		{"Deserialize transaction 2", args{Transaction2Encoded}, &Transaction2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{}
			got, err := gobenc.DeserializeTransaction(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGobEncoder_SerializeBlock(t *testing.T) {
	type args struct {
		b Block
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Serialize Block 1", args{Block1}, Block1Encoded, false},
		{"Serialize Block 2", args{Block2}, Block2Encoded, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{}
			got, err := gobenc.SerializeBlock(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SerializeBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGobEncoder_SerializeTransaction(t *testing.T) {
	type args struct {
		tx Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Serialize Transaction 1", args{Transaction1}, Transaction1Encoded, false},
		{"Serialize Transaction 2", args{Transaction2}, Transaction2Encoded, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{}
			got, err := gobenc.SerializeTransaction(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SerializeTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGobEncoder(t *testing.T) {
	tests := []struct {
		name string
		want *GobEncoder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGobEncoder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGobEncoder() = %v, want %v", got, tt.want)
			}
		})
	}
}
