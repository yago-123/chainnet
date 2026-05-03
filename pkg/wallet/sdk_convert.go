package wallet

import (
	sdkv1beta "github.com/yago-123/chainnet-sdk-go/v1beta"
	"github.com/yago-123/chainnet/pkg/kernel"
)

// These functions are needed in order to access

func KernelTransactionToSDK(tx kernel.Transaction) sdkv1beta.Transaction {
	inputs := make([]sdkv1beta.TxInput, 0, len(tx.Vin))
	for _, input := range tx.Vin {
		inputs = append(inputs, sdkv1beta.TxInput{
			Txid:      input.Txid,
			Vout:      input.Vout,
			ScriptSig: input.ScriptSig,
			PubKey:    input.PubKey,
		})
	}

	outputs := make([]sdkv1beta.TxOutput, 0, len(tx.Vout))
	for _, output := range tx.Vout {
		outputs = append(outputs, sdkv1beta.TxOutput{
			Amount:       output.Amount,
			ScriptPubKey: output.ScriptPubKey,
			PubKey:       output.PubKey,
		})
	}

	return sdkv1beta.Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: outputs,
	}
}

func SDKTransactionToKernel(tx *sdkv1beta.Transaction) *kernel.Transaction {
	inputs := make([]kernel.TxInput, 0, len(tx.Vin))
	for _, input := range tx.Vin {
		inputs = append(inputs, kernel.TxInput{
			Txid:      input.Txid,
			Vout:      input.Vout,
			ScriptSig: input.ScriptSig,
			PubKey:    input.PubKey,
		})
	}

	outputs := make([]kernel.TxOutput, 0, len(tx.Vout))
	for _, output := range tx.Vout {
		outputs = append(outputs, kernel.TxOutput{
			Amount:       output.Amount,
			ScriptPubKey: output.ScriptPubKey,
			PubKey:       output.PubKey,
		})
	}

	return &kernel.Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: outputs,
	}
}

func SDKUTXOToKernel(utxo sdkv1beta.UTXO) *kernel.UTXO {
	return &kernel.UTXO{
		TxID:   utxo.TxID,
		OutIdx: utxo.OutIdx,
		Output: kernel.TxOutput{
			Amount:       utxo.Output.Amount,
			ScriptPubKey: utxo.Output.ScriptPubKey,
			PubKey:       utxo.Output.PubKey,
		},
	}
}

func SDKUTXOsToKernel(utxos []sdkv1beta.UTXO) []*kernel.UTXO {
	ret := make([]*kernel.UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		ret = append(ret, SDKUTXOToKernel(utxo))
	}

	return ret
}
