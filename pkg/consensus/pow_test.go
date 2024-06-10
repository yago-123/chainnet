package consensus

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"math/big"
	"reflect"
	"testing"
)

func TestNewProofOfWork(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name string
		args args
		want *ProofOfWork
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProofOfWork(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProofOfWork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofOfWork_Calculate(t *testing.T) {
	type fields struct {
		target *big.Int
		cfg    *config.Config
	}
	type args struct {
		block *block.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
		want1  uint
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pow := &ProofOfWork{
				target: tt.fields.target,
				cfg:    tt.fields.cfg,
			}
			got, got1 := pow.Calculate(tt.args.block)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Calculate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Calculate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestProofOfWork_Validate(t *testing.T) {
	type fields struct {
		target *big.Int
		cfg    *config.Config
	}
	type args struct {
		block *block.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pow := &ProofOfWork{
				target: tt.fields.target,
				cfg:    tt.fields.cfg,
			}
			if got := pow.Validate(tt.args.block); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofOfWork_assembleProofData(t *testing.T) {
	type fields struct {
		target *big.Int
		cfg    *config.Config
	}
	type args struct {
		block *block.Block
		nonce uint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pow := &ProofOfWork{
				target: tt.fields.target,
				cfg:    tt.fields.cfg,
			}
			if got := pow.assembleProofData(tt.args.block, tt.args.nonce); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assembleProofData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofOfWork_hashTransactions(t *testing.T) {
	type fields struct {
		target *big.Int
		cfg    *config.Config
	}
	type args struct {
		transactions []*block.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pow := &ProofOfWork{
				target: tt.fields.target,
				cfg:    tt.fields.cfg,
			}
			if got := pow.hashTransactions(tt.args.transactions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hashTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
