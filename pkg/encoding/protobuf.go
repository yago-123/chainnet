package encoding

import (
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
	pb "github.com/yago-123/chainnet/pkg/network/protobuf"

	"google.golang.org/protobuf/proto"
)

type Protobuf struct {
}

func NewProtobufEncoder() *Protobuf {
	return &Protobuf{}
}

func (p *Protobuf) Type() string {
	return ProtoEncodingType
}

// SerializeBlock serializes a kernel.Block into a Protobuf byte array
func (p *Protobuf) SerializeBlock(b kernel.Block) ([]byte, error) {
	pbBlock, err := convertToProtobufBlock(b)
	if err != nil {
		return nil, fmt.Errorf("error converting block %x: %w", b.Hash, err)
	}
	data, err := proto.Marshal(pbBlock)
	if err != nil {
		return nil, fmt.Errorf("error serializing block %x: %w", b.Hash, err)
	}
	return data, nil
}

// DeserializeBlock deserializes a Protobuf byte array into a kernel.Block
func (p *Protobuf) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var pbBlock pb.Block
	err := proto.Unmarshal(data, &pbBlock)
	if err != nil {
		return nil, fmt.Errorf("error deserializing block: %w", err)
	}

	b, err := convertFromProtobufBlock(&pbBlock)
	if err != nil {
		return nil, fmt.Errorf("error converting block from protobuf: %w", err)
	}

	return &b, nil
}

// SerializeHeader serializes a kernel.BlockHeader into a Protobuf byte array
func (p *Protobuf) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	pbHeader := convertToProtobufBlockHeader(bh)
	data, err := proto.Marshal(pbHeader)
	if err != nil {
		return nil, fmt.Errorf("error serializing block header: %w", err)
	}
	return data, nil
}

// DeserializeHeader deserializes a Protobuf byte array into a kernel.BlockHeader
func (p *Protobuf) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	var pbHeader pb.BlockHeader
	err := proto.Unmarshal(data, &pbHeader)
	if err != nil {
		return nil, fmt.Errorf("error deserializing block header: %w", err)
	}
	bh := convertFromProtobufBlockHeader(&pbHeader)
	return bh, nil
}

// SerializeHeaders serializes a slice of kernel.BlockHeader into a Protobuf byte array.
func (p *Protobuf) SerializeHeaders(bhs []*kernel.BlockHeader) ([]byte, error) {
	var pbHeaders []*pb.BlockHeader

	for _, bh := range bhs {
		pbHeader := convertToProtobufBlockHeader(*bh)
		pbHeaders = append(pbHeaders, pbHeader)
	}

	// create a Protobuf BlockHeaders message and set headers field
	container := &pb.BlockHeaders{
		Headers: pbHeaders,
	}

	data, err := proto.Marshal(container)
	if err != nil {
		return nil, fmt.Errorf("error serializing block headers: %w", err)
	}

	return data, nil
}

// DeserializeHeaders deserializes a Protobuf byte array into a slice of kernel.BlockHeader.
func (p *Protobuf) DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error) {
	var pbHeaders pb.BlockHeaders

	if err := proto.Unmarshal(data, &pbHeaders); err != nil {
		return nil, fmt.Errorf("error deserializing block headers: %w", err)
	}

	// convert each Protobuf BlockHeader to a kernel.BlockHeader
	var bhs []*kernel.BlockHeader
	for _, pbHeader := range pbHeaders.GetHeaders() {
		bh := convertFromProtobufBlockHeader(pbHeader)
		bhs = append(bhs, bh)
	}

	return bhs, nil
}

// SerializeTransaction serializes a kernel.Transaction into a Protobuf byte array
func (p *Protobuf) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	pbTransaction := convertToProtobufTransaction(tx)
	data, err := proto.Marshal(pbTransaction)
	if err != nil {
		return nil, fmt.Errorf("error serializing transaction %x: %w", tx.ID, err)
	}
	return data, nil
}

// DeserializeTransaction deserializes a Protobuf byte array into a kernel.Transaction
func (p *Protobuf) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var pbTransaction pb.Transaction
	err := proto.Unmarshal(data, &pbTransaction)
	if err != nil {
		return nil, fmt.Errorf("error deserializing transaction: %w", err)
	}
	tx, err := convertFromProtobufTransaction(&pbTransaction)
	if err != nil {
		return nil, fmt.Errorf("error converting transaction from protobuf: %w", err)
	}
	return &tx, nil
}

// SerializeTransactions serializes a slice of kernel.Transaction into a Protobuf byte array
func (p *Protobuf) SerializeTransactions(txs []*kernel.Transaction) ([]byte, error) {
	var pbTxs []*pb.Transaction

	for _, tx := range txs {
		pbTx := convertToProtobufTransaction(*tx)
		pbTxs = append(pbTxs, pbTx)
	}

	// create a Protobuf Transactions message and set transactions field
	container := &pb.Transactions{
		Transactions: pbTxs,
	}

	data, err := proto.Marshal(container)
	if err != nil {
		return nil, fmt.Errorf("error serializing transactions: %w", err)
	}

	return data, nil
}

// DeserializeTransactions deserializes a Protobuf byte array into a slice of kernel.Transaction
func (p *Protobuf) DeserializeTransactions(data []byte) ([]*kernel.Transaction, error) {
	var pbTxs pb.Transactions

	if err := proto.Unmarshal(data, &pbTxs); err != nil {
		return nil, fmt.Errorf("error deserializing transactions: %w", err)
	}

	// convert each Protobuf Transaction to a kernel.Transaction
	var txs []*kernel.Transaction
	for _, pbTx := range pbTxs.GetTransactions() {
		tx, err := convertFromProtobufTransaction(pbTx)
		if err != nil {
			return []*kernel.Transaction{}, fmt.Errorf("error converting transaction from protobuf: %w", err)
		}

		txs = append(txs, &tx)
	}

	return txs, nil
}

// SerializeUTXO serializes a kernel.UTXO into a Protobuf byte array
func (p *Protobuf) SerializeUTXO(utxo kernel.UTXO) ([]byte, error) {
	pbUTXO := convertToProtobufUTXO(utxo)
	data, err := proto.Marshal(pbUTXO)
	if err != nil {
		return nil, fmt.Errorf("error serializing UTXO %s-%d: %w", utxo.TxID, utxo.OutIdx, err)
	}

	return data, nil
}

// DeserializeUTXO deserializes a Protobuf byte array into a kernel.UTXO
func (p *Protobuf) DeserializeUTXO(data []byte) (*kernel.UTXO, error) {
	var pbUTXO pb.UTXO
	err := proto.Unmarshal(data, &pbUTXO)
	if err != nil {
		return nil, fmt.Errorf("error deserializing UTXO: %w", err)
	}

	utxo, err := convertFromProtobufUTXO(&pbUTXO)
	if err != nil {
		return nil, fmt.Errorf("error converting UTXO from protobuf: %w", err)
	}

	return &utxo, nil
}

// SerializeUTXOs serializes a slice of kernel.UTXO into a Protobuf byte array
func (p *Protobuf) SerializeUTXOs(utxos []*kernel.UTXO) ([]byte, error) {
	var pbUtxos []*pb.UTXO

	for _, utxo := range utxos {
		pbUtxo := convertToProtobufUTXO(*utxo)
		pbUtxos = append(pbUtxos, pbUtxo)
	}

	// create a Protobuf UTXOs message and set utxos field
	container := &pb.UTXOs{
		Utxos: pbUtxos,
	}

	data, err := proto.Marshal(container)
	if err != nil {
		return nil, fmt.Errorf("error serializing UTXOs: %w", err)
	}

	return data, nil
}

// DeserializeUTXOs deserializes a Protobuf byte array into a slice of kernel.UTXO
func (p *Protobuf) DeserializeUTXOs(data []byte) ([]*kernel.UTXO, error) {
	var pbUtxos pb.UTXOs

	if err := proto.Unmarshal(data, &pbUtxos); err != nil {
		return nil, fmt.Errorf("error deserializing UTXOs: %w", err)
	}

	// convert each Protobuf UTXO to a kernel.UTXO
	var utxos []*kernel.UTXO
	for _, pbUtxo := range pbUtxos.GetUtxos() {
		utxo, err := convertFromProtobufUTXO(pbUtxo)
		if err != nil {
			return []*kernel.UTXO{}, fmt.Errorf("error converting UTXO from protobuf: %w", err)
		}
		utxos = append(utxos, &utxo)
	}

	return utxos, nil
}

func convertToProtobufBlock(b kernel.Block) (*pb.Block, error) {
	if b.Header == nil {
		return &pb.Block{}, fmt.Errorf("empty header, not safe to serialize")
	}

	return &pb.Block{
		Header:       convertToProtobufBlockHeader(*b.Header),
		Transactions: convertToProtobufTransactions(b.Transactions),
		Hash:         b.Hash,
	}, nil
}

func convertFromProtobufBlock(pbBlock *pb.Block) (kernel.Block, error) {
	txs, err := convertFromProtobufTransactions(pbBlock.GetTransactions())
	if err != nil {
		return kernel.Block{}, err
	}

	return kernel.Block{
		Header:       convertFromProtobufBlockHeader(pbBlock.GetHeader()),
		Transactions: txs,
		Hash:         pbBlock.GetHash(),
	}, nil
}

func convertToProtobufBlockHeader(bh kernel.BlockHeader) *pb.BlockHeader {
	return &pb.BlockHeader{
		Version:       bh.Version,
		PrevBlockHash: bh.PrevBlockHash,
		MerkleRoot:    bh.MerkleRoot,
		Height:        uint64(bh.Height),
		Timestamp:     bh.Timestamp,
		Target:        uint64(bh.Target),
		Nonce:         uint64(bh.Nonce),
	}
}

func convertFromProtobufBlockHeader(pbHeader *pb.BlockHeader) *kernel.BlockHeader {
	return &kernel.BlockHeader{
		Version:       pbHeader.GetVersion(),
		PrevBlockHash: pbHeader.GetPrevBlockHash(),
		MerkleRoot:    pbHeader.GetMerkleRoot(),
		Height:        uint(pbHeader.GetHeight()),
		Timestamp:     pbHeader.GetTimestamp(),
		Target:        uint(pbHeader.GetTarget()),
		Nonce:         uint(pbHeader.GetNonce()),
	}
}

func convertToProtobufTransaction(tx kernel.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:   tx.ID,
		Vin:  convertToProtobufTxInputs(tx.Vin),
		Vout: convertToProtobufTxOutputs(tx.Vout),
	}
}

func convertFromProtobufTransaction(pbTransaction *pb.Transaction) (kernel.Transaction, error) {
	txInput, err := convertFromProtobufTxInputs(pbTransaction.GetVin())
	if err != nil {
		return kernel.Transaction{}, err
	}

	txOutput, err := convertFromProtobufTxOutputs(pbTransaction.GetVout())
	if err != nil {
		return kernel.Transaction{}, err
	}

	return kernel.Transaction{
		ID:   pbTransaction.GetId(),
		Vin:  txInput,
		Vout: txOutput,
	}, nil
}

func convertToProtobufTxInput(txin kernel.TxInput) *pb.TxInput {
	return &pb.TxInput{
		Txid:      txin.Txid,
		Vout:      uint64(txin.Vout),
		ScriptSig: txin.ScriptSig,
	}
}

func convertFromProtobufTxInput(pbInput *pb.TxInput) (kernel.TxInput, error) {
	return kernel.TxInput{
		Txid:      pbInput.GetTxid(),
		Vout:      uint(pbInput.GetVout()),
		ScriptSig: pbInput.GetScriptSig(),
	}, nil
}

func convertToProtobufTxOutput(txout kernel.TxOutput) *pb.TxOutput {
	return &pb.TxOutput{
		Amount:       uint64(txout.Amount),
		ScriptPubKey: txout.ScriptPubKey,
	}
}

func convertFromProtobufTxOutput(pbOutput *pb.TxOutput) (kernel.TxOutput, error) {
	return kernel.TxOutput{
		Amount:       uint(pbOutput.GetAmount()),
		ScriptPubKey: pbOutput.GetScriptPubKey(),
	}, nil
}

func convertToProtobufTransactions(transactions []*kernel.Transaction) []*pb.Transaction {
	var pbTxs []*pb.Transaction
	for _, tx := range transactions {
		pbTxs = append(pbTxs, convertToProtobufTransaction(*tx))
	}
	return pbTxs
}

func convertFromProtobufTransactions(pbs []*pb.Transaction) ([]*kernel.Transaction, error) {
	var txs []*kernel.Transaction
	for _, pb := range pbs {
		tx, err := convertFromProtobufTransaction(pb)
		if err != nil {
			return []*kernel.Transaction{}, err
		}
		txs = append(txs, &tx)
	}
	return txs, nil
}

func convertToProtobufTxInputs(inputs []kernel.TxInput) []*pb.TxInput {
	var pbInputs []*pb.TxInput
	for _, in := range inputs {
		pbInputs = append(pbInputs, convertToProtobufTxInput(in))
	}
	return pbInputs
}

func convertFromProtobufTxInputs(pbs []*pb.TxInput) ([]kernel.TxInput, error) {
	var inputs []kernel.TxInput
	for _, pb := range pbs {
		txInput, err := convertFromProtobufTxInput(pb)
		if err != nil {
			return []kernel.TxInput{}, err
		}
		inputs = append(inputs, txInput)
	}
	return inputs, nil
}

func convertToProtobufTxOutputs(outputs []kernel.TxOutput) []*pb.TxOutput {
	var pbOutputs []*pb.TxOutput
	for _, out := range outputs {
		pbOutputs = append(pbOutputs, convertToProtobufTxOutput(out))
	}
	return pbOutputs
}

func convertFromProtobufTxOutputs(pbs []*pb.TxOutput) ([]kernel.TxOutput, error) {
	var outputs []kernel.TxOutput

	for _, pb := range pbs {
		txOutput, err := convertFromProtobufTxOutput(pb)
		if err != nil {
			return []kernel.TxOutput{}, err
		}
		outputs = append(outputs, txOutput)
	}
	return outputs, nil
}

func convertToProtobufUTXO(utxo kernel.UTXO) *pb.UTXO {
	return &pb.UTXO{
		Txid:   utxo.TxID,
		Vout:   uint64(utxo.OutIdx),
		Output: convertToProtobufTxOutput(utxo.Output),
	}
}

func convertFromProtobufUTXO(pbUTXO *pb.UTXO) (kernel.UTXO, error) {
	txOutput, err := convertFromProtobufTxOutput(pbUTXO.GetOutput())
	if err != nil {
		return kernel.UTXO{}, err
	}

	return kernel.UTXO{
		TxID:   pbUTXO.GetTxid(),
		OutIdx: uint(pbUTXO.GetVout()),
		Output: txOutput,
	}, nil
}
