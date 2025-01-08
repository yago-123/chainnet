package wallet

import (
	"errors"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
)

// GenerateInputs set up the inputs for the transaction and returns the total balance of the UTXOs that are going to be
// spent in the transaction
func GenerateInputs(utxos []*kernel.UTXO, targetAmount uint) ([]kernel.TxInput, uint, error) {
	// for now simple FIFO method, first outputs are the first to be spent
	balance := uint(0)
	inputs := []kernel.TxInput{}

	for _, utxo := range utxos {
		balance += utxo.Output.Amount
		inputs = append(inputs, kernel.NewInput(utxo.TxID, utxo.OutIdx, "", utxo.Output.PubKey))
		if balance >= targetAmount {
			return inputs, balance, nil
		}
	}

	return []kernel.TxInput{}, balance, errors.New("not enough funds to perform the transaction")
}

// GenerateOutputs set up the outputs for the transaction
func GenerateOutputs(scriptType script.ScriptType, targetAmount, txFee, totalBalance uint, receiver, changeReceiver string) []kernel.TxOutput {
	change := totalBalance - txFee - targetAmount

	txOutput := []kernel.TxOutput{}
	txOutput = append(txOutput, kernel.NewOutput(targetAmount, scriptType, receiver))

	// add output corresponding to the spare changeType
	if change > 0 {
		txOutput = append(txOutput, kernel.NewOutput(totalBalance-txFee-targetAmount, scriptType, changeReceiver))
	}

	return txOutput
}
