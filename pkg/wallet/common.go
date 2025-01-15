package wallet

import (
	"errors"
	"fmt"

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
// - targetAmounts: the amount that is going to be sent to the receivers
// - addresses: were the outputs are sent
// - txFee: the amount that is going to be used as a fee for the transaction
// - totalBalance: the total balance of the UTXOs that are going to be spent
// - changeReceiverPubKey: the public key of the change output (can't paste the address because will change based
// on the payment type (scriptType)
// - changeReceiverVersion: the version of the change output
func GenerateOutputs(scriptType script.ScriptType, targetAmounts []uint, addresses [][]byte, txFee, totalBalance uint, changeReceiverPubKey []byte, changeReceiverVersion byte) ([]kernel.TxOutput, error) {
	txOutput := []kernel.TxOutput{}

	// check if the target amount and the addresses have the same length
	if len(targetAmounts) != len(addresses) {
		return []kernel.TxOutput{}, fmt.Errorf("target amounts (len=%d) and addresses (len=%d) must have the same length", len(targetAmounts), len(addresses))
	}

	// calculate the total amount that is going to be sent to the addresses
	totalTargetAmount := uint(0)
	for _, amount := range targetAmounts {
		totalTargetAmount += amount
	}
	if totalTargetAmount+txFee > totalBalance {
		return []kernel.TxOutput{}, errors.New("not enough funds to perform the transaction")
	}

	// calculate the spare change
	change := totalBalance - txFee - totalTargetAmount

	// generate one output for each receiver
	for i := range addresses {
		txOutput = append(txOutput, kernel.NewOutput(targetAmounts[i], scriptType, string(addresses[i])))
	}

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

		txOutput = append(txOutput, kernel.NewOutput(totalBalance-txFee-totalTargetAmount, scriptType, changeAddress))
	}

	return txOutput, nil
}
