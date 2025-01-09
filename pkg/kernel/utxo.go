package kernel

import (
	"bytes"
	"fmt"
)

// UTXO represents the unspent transaction output
type UTXO struct {
	TxID   []byte
	OutIdx uint
	Output TxOutput
}

// EqualInput checks if the input is the same as the given input
func (utxo *UTXO) EqualInput(input TxInput) bool {
	return bytes.Equal(utxo.TxID, input.Txid) && utxo.OutIdx == input.Vout
}

// UniqueKey represents the unique key for the UTXO. Method used for mapping UTXOs and inputs via this unique key
func (utxo *UTXO) UniqueKey() string {
	return fmt.Sprintf("%x-%d", utxo.TxID, utxo.OutIdx)
}

// GetAmount returns the balance value contained in the UTXO ($$$)
func (utxo *UTXO) GetAmount() uint {
	return utxo.Output.Amount
}
