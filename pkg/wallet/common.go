package wallet

import (
	"errors"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"
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

// GenerateOutputs set up the outputs for the transaction and handle the change if necessary
// Arguments:
// - scriptType: the type of script that is going to be used in the outputs
// - targetAmount: the amount that is going to be sent to the receiver
// - txFee: the amount that is going to be used as a fee for the transaction
// - totalBalance: the total balance of the UTXOs that are going to be spent
// - receiver: the address of the main output
// - changeReceiverPubKey: the public key of the change output (can't paste the address because will change based
// on the payment type (scriptType)
// - changeReceiverVersion: the version of the change output
func GenerateOutputs(scriptType script.ScriptType, targetAmount, txFee, totalBalance uint, receiver, changeReceiverPubKey []byte, changeReceiverVersion byte) ([]kernel.TxOutput, error) {
	change := totalBalance - txFee - targetAmount

	txOutput := []kernel.TxOutput{}
	txOutput = append(txOutput, kernel.NewOutput(targetAmount, scriptType, string(receiver)))

	// add output corresponding to the spare changeType. The change address will be calculated based on the scriptType
	// desired, in order to do so we calculate the address based on the public key of the change receiver
	if change > 0 {
		changeAddress := ""
		// calculate the address for P2PK
		if scriptType == script.P2PK {
			changeAddress = string(changeReceiverPubKey)
		}

		// calculate the address for P2PKH
		if scriptType == script.P2PKH {
			changeAddressArray, err := util_p2pkh.GenerateP2PKHAddrFromPubKey(changeReceiverPubKey, changeReceiverVersion)
			if err != nil {
				return []kernel.TxOutput{}, err
			}

			changeAddress = string(changeAddressArray)
		}

		txOutput = append(txOutput, kernel.NewOutput(totalBalance-txFee-targetAmount, scriptType, changeAddress))
	}

	return txOutput, nil
}
