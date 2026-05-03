package v1beta

import "github.com/yago-123/chainnet-sdk-go/v1beta/generated"

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

type BlockHeader struct {
	Version       []byte
	PrevBlockHash []byte
	MerkleRoot    []byte
	Height        uint
	Timestamp     int64
	Target        uint
	Nonce         uint
}

type ChainTip = generated.ChainTip
type ErrorResponse = generated.ErrorResponse

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type TxInput struct {
	Txid      []byte
	Vout      uint
	ScriptSig string
	PubKey    string
}

type TxOutput struct {
	Amount       uint
	ScriptPubKey string
	PubKey       string
}

type UTXO struct {
	TxID   []byte
	OutIdx uint
	Output TxOutput
}

func (utxo UTXO) Amount() uint {
	return utxo.Output.Amount
}

type HeaderOrder = generated.GetHeadersParamsOrder

const (
	HeaderOrderNewestFirst HeaderOrder = generated.NewestFirst
	HeaderOrderOldestFirst HeaderOrder = generated.OldestFirst
)

type GetHeadersOptions struct {
	Limit *int
	Order *HeaderOrder
}
