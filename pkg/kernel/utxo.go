package kernel

import "bytes"

// UnspentOutput represents the unspent transaction output
type UnspentOutput struct {
	TxID   []byte
	OutIdx uint
	Output TxOutput
}

// EqualInput checks if the input is the same as the given input
func (utxo *UnspentOutput) EqualInput(input TxInput) bool {
	return bytes.Equal(utxo.TxID, input.Txid) && utxo.OutIdx == input.Vout
}
