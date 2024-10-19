package encoding

import (
	"github.com/yago-123/chainnet/pkg/kernel"
	pb "github.com/yago-123/chainnet/pkg/p2p/protobuf"
	"encoding/hex"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Protobuf struct {
}

func NewProtobufEncoder() *Protobuf {
	return &Protobuf{}
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
	var pbHeaders pb.BlockHeaders // Adjust to your Protobuf message type.

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
		PubKey:    fmt.Sprintf("%x", txin.PubKey),
	}
}

func convertFromProtobufTxInput(pbInput *pb.TxInput) (kernel.TxInput, error) {
	decodedPubKey, err := hex.DecodeString(pbInput.GetPubKey())
	if err != nil {
		return kernel.TxInput{}, fmt.Errorf("error decoding pubkey %s: %w", pbInput.GetPubKey(), err)
	}

	return kernel.TxInput{
		Txid:      pbInput.GetTxid(),
		Vout:      uint(pbInput.GetVout()),
		ScriptSig: pbInput.GetScriptSig(),
		PubKey:    string(decodedPubKey),
	}, nil
}

func convertToProtobufTxOutput(txout kernel.TxOutput) *pb.TxOutput {
	return &pb.TxOutput{
		Amount:       uint64(txout.Amount),
		ScriptPubKey: txout.ScriptPubKey,
		PubKey:       fmt.Sprintf("%x", txout.PubKey),
	}
}

func convertFromProtobufTxOutput(pbOutput *pb.TxOutput) (kernel.TxOutput, error) {
	decodedPubKey, err := hex.DecodeString(pbOutput.GetPubKey())
	if err != nil {
		return kernel.TxOutput{}, fmt.Errorf("error decoding pubkey %s: %w", pbOutput.GetPubKey(), err)
	}

	return kernel.TxOutput{
		Amount:       uint(pbOutput.GetAmount()),
		ScriptPubKey: pbOutput.GetScriptPubKey(),
		PubKey:       string(decodedPubKey),
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
