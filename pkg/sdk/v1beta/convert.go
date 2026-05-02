package v1beta

import (
	"encoding/hex"
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/sdk/v1beta/generated"
)

func encodeHash(value []byte) string {
	return hex.EncodeToString(value)
}

func blockFromGenerated(block generated.Block) (*kernel.Block, error) {
	header, err := headerFromGenerated(block.Header)
	if err != nil {
		return nil, err
	}

	transactions, err := transactionsFromGenerated(block.Transactions)
	if err != nil {
		return nil, err
	}

	hash, err := decodeHexField("hash", block.Hash)
	if err != nil {
		return nil, err
	}

	return &kernel.Block{
		Header:       header,
		Transactions: transactions,
		Hash:         hash,
	}, nil
}

func headerFromGenerated(header generated.BlockHeader) (*kernel.BlockHeader, error) {
	version, err := decodeHexField("version", header.Version)
	if err != nil {
		return nil, err
	}
	prevBlockHash, err := decodeHexField("prev_block_hash", header.PrevBlockHash)
	if err != nil {
		return nil, err
	}
	merkleRoot, err := decodeHexField("merkle_root", header.MerkleRoot)
	if err != nil {
		return nil, err
	}
	height, err := intToUint("height", header.Height)
	if err != nil {
		return nil, err
	}
	target, err := intToUint("target", header.Target)
	if err != nil {
		return nil, err
	}
	nonce, err := intToUint("nonce", header.Nonce)
	if err != nil {
		return nil, err
	}

	return &kernel.BlockHeader{
		Version:       version,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    merkleRoot,
		Height:        height,
		Timestamp:     header.Timestamp,
		Target:        target,
		Nonce:         nonce,
	}, nil
}

func headersFromGenerated(headers []generated.BlockHeader) ([]*kernel.BlockHeader, error) {
	ret := make([]*kernel.BlockHeader, 0, len(headers))
	for _, header := range headers {
		converted, err := headerFromGenerated(header)
		if err != nil {
			return nil, err
		}

		ret = append(ret, converted)
	}

	return ret, nil
}

func transactionToGenerated(tx kernel.Transaction) (generated.Transaction, error) {
	inputs, err := txInputsToGenerated(tx.Vin)
	if err != nil {
		return generated.Transaction{}, err
	}
	outputs, err := txOutputsToGenerated(tx.Vout)
	if err != nil {
		return generated.Transaction{}, err
	}

	return generated.Transaction{
		Id:   encodeHash(tx.ID),
		Vin:  inputs,
		Vout: outputs,
	}, nil
}

func transactionFromGenerated(tx generated.Transaction) (*kernel.Transaction, error) {
	id, err := decodeHexField("id", tx.Id)
	if err != nil {
		return nil, err
	}

	inputs, err := txInputsFromGenerated(tx.Vin)
	if err != nil {
		return nil, err
	}

	outputs, err := txOutputsFromGenerated(tx.Vout)
	if err != nil {
		return nil, err
	}

	return &kernel.Transaction{
		ID:   id,
		Vin:  inputs,
		Vout: outputs,
	}, nil
}

func transactionsFromGenerated(txs []generated.Transaction) ([]*kernel.Transaction, error) {
	ret := make([]*kernel.Transaction, 0, len(txs))
	for _, tx := range txs {
		converted, err := transactionFromGenerated(tx)
		if err != nil {
			return nil, err
		}

		ret = append(ret, converted)
	}

	return ret, nil
}

func txInputsToGenerated(inputs []kernel.TxInput) ([]generated.TxInput, error) {
	ret := make([]generated.TxInput, 0, len(inputs))
	for _, input := range inputs {
		vout, err := uintToInt("vout", input.Vout)
		if err != nil {
			return nil, err
		}

		ret = append(ret, generated.TxInput{
			Txid:      encodeHash(input.Txid),
			Vout:      vout,
			ScriptSig: input.ScriptSig,
			PubKey:    input.PubKey,
		})
	}

	return ret, nil
}

func txInputsFromGenerated(inputs []generated.TxInput) ([]kernel.TxInput, error) {
	ret := make([]kernel.TxInput, 0, len(inputs))
	for _, input := range inputs {
		txID, err := decodeHexField("txid", input.Txid)
		if err != nil {
			return nil, err
		}
		vout, err := intToUint("vout", input.Vout)
		if err != nil {
			return nil, err
		}

		ret = append(ret, kernel.TxInput{
			Txid:      txID,
			Vout:      vout,
			ScriptSig: input.ScriptSig,
			PubKey:    input.PubKey,
		})
	}

	return ret, nil
}

func txOutputsToGenerated(outputs []kernel.TxOutput) ([]generated.TxOutput, error) {
	ret := make([]generated.TxOutput, 0, len(outputs))
	for _, output := range outputs {
		amount, err := uintToInt("amount", output.Amount)
		if err != nil {
			return nil, err
		}

		ret = append(ret, generated.TxOutput{
			Amount:       amount,
			ScriptPubKey: output.ScriptPubKey,
			PubKey:       output.PubKey,
		})
	}

	return ret, nil
}

func txOutputsFromGenerated(outputs []generated.TxOutput) ([]kernel.TxOutput, error) {
	ret := make([]kernel.TxOutput, 0, len(outputs))
	for _, output := range outputs {
		amount, err := intToUint("amount", output.Amount)
		if err != nil {
			return nil, err
		}

		ret = append(ret, kernel.TxOutput{
			Amount:       amount,
			ScriptPubKey: output.ScriptPubKey,
			PubKey:       output.PubKey,
		})
	}

	return ret, nil
}

func utxosFromGenerated(utxos []generated.UTXO) ([]*kernel.UTXO, error) {
	ret := make([]*kernel.UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		txID, err := decodeHexField("txid", utxo.Txid)
		if err != nil {
			return nil, err
		}
		outIdx, err := intToUint("vout", utxo.Vout)
		if err != nil {
			return nil, err
		}
		output, err := txOutputFromGenerated(utxo.Output)
		if err != nil {
			return nil, err
		}

		ret = append(ret, &kernel.UTXO{
			TxID:   txID,
			OutIdx: outIdx,
			Output: output,
		})
	}

	return ret, nil
}

func txOutputFromGenerated(output generated.TxOutput) (kernel.TxOutput, error) {
	amount, err := intToUint("amount", output.Amount)
	if err != nil {
		return kernel.TxOutput{}, err
	}

	return kernel.TxOutput{
		Amount:       amount,
		ScriptPubKey: output.ScriptPubKey,
		PubKey:       output.PubKey,
	}, nil
}

func decodeHexField(field string, value string) ([]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("error decoding %s: %w", field, err)
	}

	return decoded, nil
}

func intToUint(field string, value int) (uint, error) {
	if value < 0 {
		return 0, fmt.Errorf("%s must be non-negative", field)
	}

	return uint(value), nil
}

func uintToInt(field string, value uint) (int, error) {
	maxInt := uint(^uint(0) >> 1)
	if value > maxInt {
		return 0, fmt.Errorf("%s exceeds int max", field)
	}

	return int(value), nil
}
