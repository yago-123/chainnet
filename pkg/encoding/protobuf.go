package encoding

import (
	pb "chainnet/pkg/chain/p2p/protobuf"
	"chainnet/pkg/kernel"
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
	pbBlock := convertToProtobufBlock(b)
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
	b := convertFromProtobufBlock(&pbBlock)
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
	tx := convertFromProtobufTransaction(&pbTransaction)
	return &tx, nil
}

func convertToProtobufBlock(b kernel.Block) *pb.Block {
	return &pb.Block{
		Header:       convertToProtobufBlockHeader(*b.Header),
		Transactions: convertToProtobufTransactions(b.Transactions),
		Hash:         b.Hash,
	}
}

func convertFromProtobufBlock(pbBlock *pb.Block) kernel.Block {
	return kernel.Block{
		Header:       convertFromProtobufBlockHeader(pbBlock.GetHeader()),
		Transactions: convertFromProtobufTransactions(pbBlock.GetTransactions()),
		Hash:         pbBlock.GetHash(),
	}
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

func convertFromProtobufTransaction(pbTransaction *pb.Transaction) kernel.Transaction {
	return kernel.Transaction{
		ID:   pbTransaction.GetId(),
		Vin:  convertFromProtobufTxInputs(pbTransaction.GetVin()),
		Vout: convertFromProtobufTxOutputs(pbTransaction.GetVout()),
	}
}

func convertToProtobufTxInput(txin kernel.TxInput) *pb.TxInput {
	return &pb.TxInput{
		Txid:      txin.Txid,
		Vout:      uint64(txin.Vout),
		ScriptSig: txin.ScriptSig,
		PubKey:    txin.PubKey,
	}
}

func convertFromProtobufTxInput(pbInput *pb.TxInput) kernel.TxInput {
	return kernel.TxInput{
		Txid:      pbInput.GetTxid(),
		Vout:      uint(pbInput.GetVout()),
		ScriptSig: pbInput.GetScriptSig(),
		PubKey:    pbInput.GetPubKey(),
	}
}

func convertToProtobufTxOutput(txout kernel.TxOutput) *pb.TxOutput {
	return &pb.TxOutput{
		Amount:       uint64(txout.Amount),
		ScriptPubKey: txout.ScriptPubKey,
		PubKey:       txout.PubKey,
	}
}

func convertFromProtobufTxOutput(pbOutput *pb.TxOutput) kernel.TxOutput {
	return kernel.TxOutput{
		Amount:       uint(pbOutput.GetAmount()),
		ScriptPubKey: pbOutput.GetScriptPubKey(),
		PubKey:       pbOutput.GetPubKey(),
	}
}

func convertToProtobufTransactions(transactions []*kernel.Transaction) []*pb.Transaction {
	var pbTxs []*pb.Transaction
	for _, tx := range transactions {
		pbTxs = append(pbTxs, convertToProtobufTransaction(*tx))
	}
	return pbTxs
}

func convertFromProtobufTransactions(pbs []*pb.Transaction) []*kernel.Transaction {
	var txs []*kernel.Transaction
	for _, pb := range pbs {
		tx := convertFromProtobufTransaction(pb)
		txs = append(txs, &tx)
	}
	return txs
}

func convertToProtobufTxInputs(inputs []kernel.TxInput) []*pb.TxInput {
	var pbInputs []*pb.TxInput
	for _, in := range inputs {
		pbInputs = append(pbInputs, convertToProtobufTxInput(in))
	}
	return pbInputs
}

func convertFromProtobufTxInputs(pbs []*pb.TxInput) []kernel.TxInput {
	var inputs []kernel.TxInput
	for _, pb := range pbs {
		inputs = append(inputs, convertFromProtobufTxInput(pb))
	}
	return inputs
}

func convertToProtobufTxOutputs(outputs []kernel.TxOutput) []*pb.TxOutput {
	var pbOutputs []*pb.TxOutput
	for _, out := range outputs {
		pbOutputs = append(pbOutputs, convertToProtobufTxOutput(out))
	}
	return pbOutputs
}

func convertFromProtobufTxOutputs(pbs []*pb.TxOutput) []kernel.TxOutput {
	var outputs []kernel.TxOutput
	for _, pb := range pbs {
		outputs = append(outputs, convertFromProtobufTxOutput(pb))
	}
	return outputs
}
