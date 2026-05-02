package encoding

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
)

type JSON struct {
}

// Notice: added these structs to avoid polluting the core components. This might change in the future
type jsonTransaction struct {
	ID   string         `json:"id"`
	Vin  []jsonTxInput  `json:"vin"`
	Vout []jsonTxOutput `json:"vout"`
}

type jsonTxInput struct {
	Txid      string `json:"txid"`
	Vout      uint   `json:"vout"`
	ScriptSig string `json:"script_sig"`
	PubKey    string `json:"pub_key"`
}

type jsonTxOutput struct {
	Amount       uint   `json:"amount"`
	ScriptPubKey string `json:"script_pub_key"`
	PubKey       string `json:"pub_key"`
}

type jsonUTXO struct {
	TxID   string       `json:"txid"`
	OutIdx uint         `json:"vout"`
	Output jsonTxOutput `json:"output"`
}

type jsonBlockHeader struct {
	Version       string `json:"version"`
	PrevBlockHash string `json:"prev_block_hash"`
	MerkleRoot    string `json:"merkle_root"`
	Height        uint   `json:"height"`
	Timestamp     int64  `json:"timestamp"`
	Target        uint   `json:"target"`
	Nonce         uint   `json:"nonce"`
}

type jsonBlock struct {
	Header       *jsonBlockHeader  `json:"header"`
	Transactions []jsonTransaction `json:"transactions"`
	Hash         string            `json:"hash"`
}

func NewJSONEncoder() *JSON {
	return &JSON{}
}

func (j *JSON) Type() string {
	return JSONEncodingType
}

func (j *JSON) SerializeBlock(b kernel.Block) ([]byte, error) {
	return json.Marshal(convertToJSONBlock(b))
}

func (j *JSON) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var block jsonBlock
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("error deserializing block: %w", err)
	}

	ret, err := convertFromJSONBlock(block)
	if err != nil {
		return nil, fmt.Errorf("error converting block from JSON: %w", err)
	}

	return &ret, nil
}

func (j *JSON) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	return json.Marshal(convertToJSONBlockHeader(bh))
}

func (j *JSON) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	var header jsonBlockHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("error deserializing block header: %w", err)
	}

	ret, err := convertFromJSONBlockHeader(header)
	if err != nil {
		return nil, fmt.Errorf("error converting block header from JSON: %w", err)
	}

	return ret, nil
}

func (j *JSON) SerializeHeaders(bhs []*kernel.BlockHeader) ([]byte, error) {
	headers := make([]jsonBlockHeader, 0, len(bhs))
	for _, header := range bhs {
		headers = append(headers, convertToJSONBlockHeader(*header))
	}

	return json.Marshal(headers)
}

func (j *JSON) DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error) {
	var headers []jsonBlockHeader
	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("error deserializing block headers: %w", err)
	}

	ret := make([]*kernel.BlockHeader, 0, len(headers))
	for _, header := range headers {
		converted, err := convertFromJSONBlockHeader(header)
		if err != nil {
			return nil, fmt.Errorf("error converting block header from JSON: %w", err)
		}

		ret = append(ret, converted)
	}

	return ret, nil
}

func (j *JSON) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	return json.Marshal(convertToJSONTransaction(tx))
}

func (j *JSON) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var tx jsonTransaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("error deserializing transaction: %w", err)
	}

	ret, err := convertFromJSONTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("error converting transaction from JSON: %w", err)
	}

	return &ret, nil
}

func (j *JSON) SerializeTransactions(txs []*kernel.Transaction) ([]byte, error) {
	transactions := make([]jsonTransaction, 0, len(txs))
	for _, tx := range txs {
		transactions = append(transactions, convertToJSONTransaction(*tx))
	}

	return json.Marshal(transactions)
}

func (j *JSON) DeserializeTransactions(data []byte) ([]*kernel.Transaction, error) {
	var txs []jsonTransaction
	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, fmt.Errorf("error deserializing transactions: %w", err)
	}

	ret := make([]*kernel.Transaction, 0, len(txs))
	for _, tx := range txs {
		converted, err := convertFromJSONTransaction(tx)
		if err != nil {
			return nil, fmt.Errorf("error converting transaction from JSON: %w", err)
		}

		ret = append(ret, &converted)
	}

	return ret, nil
}

func (j *JSON) SerializeUTXO(utxo kernel.UTXO) ([]byte, error) {
	return json.Marshal(convertToJSONUTXO(utxo))
}

func (j *JSON) DeserializeUTXO(data []byte) (*kernel.UTXO, error) {
	var utxo jsonUTXO
	if err := json.Unmarshal(data, &utxo); err != nil {
		return nil, fmt.Errorf("error deserializing UTXO: %w", err)
	}

	ret, err := convertFromJSONUTXO(utxo)
	if err != nil {
		return nil, fmt.Errorf("error converting UTXO from JSON: %w", err)
	}

	return &ret, nil
}

func (j *JSON) SerializeUTXOs(utxos []*kernel.UTXO) ([]byte, error) {
	jsonUTXOs := make([]jsonUTXO, 0, len(utxos))
	for _, utxo := range utxos {
		jsonUTXOs = append(jsonUTXOs, convertToJSONUTXO(*utxo))
	}

	return json.Marshal(jsonUTXOs)
}

func (j *JSON) DeserializeUTXOs(data []byte) ([]*kernel.UTXO, error) {
	var utxos []jsonUTXO
	if err := json.Unmarshal(data, &utxos); err != nil {
		return nil, fmt.Errorf("error deserializing UTXOs: %w", err)
	}

	ret := make([]*kernel.UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		converted, err := convertFromJSONUTXO(utxo)
		if err != nil {
			return nil, fmt.Errorf("error converting UTXO from JSON: %w", err)
		}

		ret = append(ret, &converted)
	}

	return ret, nil
}

func (j *JSON) SerializeBool(b bool) ([]byte, error) {
	return json.Marshal(b)
}

func (j *JSON) DeserializeBool(data []byte) (bool, error) {
	var ret bool
	if err := json.Unmarshal(data, &ret); err != nil {
		return false, fmt.Errorf("error deserializing bool: %w", err)
	}

	return ret, nil
}

func convertToJSONBlock(b kernel.Block) jsonBlock {
	var header *jsonBlockHeader
	if b.Header != nil {
		convertedHeader := convertToJSONBlockHeader(*b.Header)
		header = &convertedHeader
	}

	return jsonBlock{
		Header:       header,
		Transactions: convertToJSONTransactions(b.Transactions),
		Hash:         hex.EncodeToString(b.Hash),
	}
}

func convertFromJSONBlock(block jsonBlock) (kernel.Block, error) {
	var header *kernel.BlockHeader
	if block.Header != nil {
		convertedHeader, err := convertFromJSONBlockHeader(*block.Header)
		if err != nil {
			return kernel.Block{}, err
		}
		header = convertedHeader
	}

	txs, err := convertFromJSONTransactions(block.Transactions)
	if err != nil {
		return kernel.Block{}, err
	}

	hash, err := decodeHexField("hash", block.Hash)
	if err != nil {
		return kernel.Block{}, err
	}

	return kernel.Block{
		Header:       header,
		Transactions: txs,
		Hash:         hash,
	}, nil
}

func convertToJSONBlockHeader(bh kernel.BlockHeader) jsonBlockHeader {
	return jsonBlockHeader{
		Version:       hex.EncodeToString(bh.Version),
		PrevBlockHash: hex.EncodeToString(bh.PrevBlockHash),
		MerkleRoot:    hex.EncodeToString(bh.MerkleRoot),
		Height:        bh.Height,
		Timestamp:     bh.Timestamp,
		Target:        bh.Target,
		Nonce:         bh.Nonce,
	}
}

func convertFromJSONBlockHeader(header jsonBlockHeader) (*kernel.BlockHeader, error) {
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

	return &kernel.BlockHeader{
		Version:       version,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    merkleRoot,
		Height:        header.Height,
		Timestamp:     header.Timestamp,
		Target:        header.Target,
		Nonce:         header.Nonce,
	}, nil
}

func convertToJSONTransaction(tx kernel.Transaction) jsonTransaction {
	return jsonTransaction{
		ID:   hex.EncodeToString(tx.ID),
		Vin:  convertToJSONTxInputs(tx.Vin),
		Vout: convertToJSONTxOutputs(tx.Vout),
	}
}

func convertFromJSONTransaction(tx jsonTransaction) (kernel.Transaction, error) {
	id, err := decodeHexField("id", tx.ID)
	if err != nil {
		return kernel.Transaction{}, err
	}

	inputs, err := convertFromJSONTxInputs(tx.Vin)
	if err != nil {
		return kernel.Transaction{}, err
	}

	outputs := convertFromJSONTxOutputs(tx.Vout)

	return kernel.Transaction{
		ID:   id,
		Vin:  inputs,
		Vout: outputs,
	}, nil
}

func convertToJSONTxInput(input kernel.TxInput) jsonTxInput {
	return jsonTxInput{
		Txid:      hex.EncodeToString(input.Txid),
		Vout:      input.Vout,
		ScriptSig: input.ScriptSig,
		PubKey:    input.PubKey,
	}
}

func convertFromJSONTxInput(input jsonTxInput) (kernel.TxInput, error) {
	txID, err := decodeHexField("txid", input.Txid)
	if err != nil {
		return kernel.TxInput{}, err
	}

	return kernel.TxInput{
		Txid:      txID,
		Vout:      input.Vout,
		ScriptSig: input.ScriptSig,
		PubKey:    input.PubKey,
	}, nil
}

func convertToJSONTxOutput(output kernel.TxOutput) jsonTxOutput {
	return jsonTxOutput{
		Amount:       output.Amount,
		ScriptPubKey: output.ScriptPubKey,
		PubKey:       output.PubKey,
	}
}

func convertFromJSONTxOutput(output jsonTxOutput) kernel.TxOutput {
	return kernel.TxOutput{
		Amount:       output.Amount,
		ScriptPubKey: output.ScriptPubKey,
		PubKey:       output.PubKey,
	}
}

func convertToJSONUTXO(utxo kernel.UTXO) jsonUTXO {
	return jsonUTXO{
		TxID:   hex.EncodeToString(utxo.TxID),
		OutIdx: utxo.OutIdx,
		Output: convertToJSONTxOutput(utxo.Output),
	}
}

func convertFromJSONUTXO(utxo jsonUTXO) (kernel.UTXO, error) {
	txID, err := decodeHexField("txid", utxo.TxID)
	if err != nil {
		return kernel.UTXO{}, err
	}

	return kernel.UTXO{
		TxID:   txID,
		OutIdx: utxo.OutIdx,
		Output: convertFromJSONTxOutput(utxo.Output),
	}, nil
}

func convertToJSONTransactions(txs []*kernel.Transaction) []jsonTransaction {
	transactions := make([]jsonTransaction, 0, len(txs))
	for _, tx := range txs {
		transactions = append(transactions, convertToJSONTransaction(*tx))
	}
	return transactions
}

func convertFromJSONTransactions(txs []jsonTransaction) ([]*kernel.Transaction, error) {
	ret := make([]*kernel.Transaction, 0, len(txs))
	for _, tx := range txs {
		converted, err := convertFromJSONTransaction(tx)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &converted)
	}
	return ret, nil
}

func convertToJSONTxInputs(inputs []kernel.TxInput) []jsonTxInput {
	ret := make([]jsonTxInput, 0, len(inputs))
	for _, input := range inputs {
		ret = append(ret, convertToJSONTxInput(input))
	}
	return ret
}

func convertFromJSONTxInputs(inputs []jsonTxInput) ([]kernel.TxInput, error) {
	ret := make([]kernel.TxInput, 0, len(inputs))
	for _, input := range inputs {
		converted, err := convertFromJSONTxInput(input)
		if err != nil {
			return nil, err
		}
		ret = append(ret, converted)
	}
	return ret, nil
}

func convertToJSONTxOutputs(outputs []kernel.TxOutput) []jsonTxOutput {
	ret := make([]jsonTxOutput, 0, len(outputs))
	for _, output := range outputs {
		ret = append(ret, convertToJSONTxOutput(output))
	}
	return ret
}

func convertFromJSONTxOutputs(outputs []jsonTxOutput) []kernel.TxOutput {
	ret := make([]kernel.TxOutput, 0, len(outputs))
	for _, output := range outputs {
		ret = append(ret, convertFromJSONTxOutput(output))
	}
	return ret
}

func decodeHexField(field string, value string) ([]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("error decoding %s: %w", field, err)
	}

	return decoded, nil
}
