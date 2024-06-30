package wallet

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"errors"

	base58 "github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	version    []byte
	PrivateKey []byte
	PublicKey  []byte

	validator consensus.LightValidator
	signer    sign.Signature
	hasher    hash.Hashing
}

func (w *Wallet) ID() string {
	return string(w.hasher.Hash(w.PublicKey))
}

func NewWallet(version []byte, validator consensus.LightValidator, signer sign.Signature, hasher hash.Hashing) (*Wallet, error) {
	privateKey, publicKey, err := signer.NewKeyPair()
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		validator:  validator,
		signer:     signer,
		hasher:     hasher,
		version:    version,
	}, nil
}

// GetAddress returns one wallet address
// todo() implement hierarchically deterministic HD wallet
func (w *Wallet) GetAddress() []byte {
	// hash the public key
	pubKeyHash := w.hasher.Hash(w.PublicKey)

	// add the version to the hashed public key in order to hash again and obtain the checksum
	versionedPayload := append(w.version, pubKeyHash...) //nolint:gocritic // we need to append the version to the payload
	checksum := w.hasher.Hash(versionedPayload)

	// return the base58 of the versioned payload and the checksum
	payload := append(versionedPayload, checksum...)
	return []byte(base58.Encode(payload))
}

// SendTransaction creates a transaction and broadcasts it to the network
func (w *Wallet) SendTransaction(to string, targetAmount uint, txFee uint, utxos []*kernel.UnspentOutput) (*kernel.Transaction, error) {
	// create the inputs necessary for the transaction
	inputs, totalBalance, err := generateInputs(utxos, targetAmount+txFee)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// create the outputs necessary for the transaction
	outputs, err := generateOutputs(targetAmount, txFee, totalBalance, to, string(w.PublicKey))
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// generate transaction
	tx := kernel.NewTransaction(
		inputs,
		outputs,
	)

	// unlock the funds from the UTXOs
	tx, err = w.UnlockTxFunds(tx)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// perform simple validations (light validator) before broadcasting the transaction
	if err = w.validator.ValidateTxLight(tx); err != nil {
		return &kernel.Transaction{}, err
	}

	return broadcastTransaction(tx)
}

// UnlockTxFunds take a tx that is being built and unlocks the UTXOs from which the input funds are going to
// be used
func (w *Wallet) UnlockTxFunds(tx *kernel.Transaction) (*kernel.Transaction, error) {

	// todo() for now, this only applies to P2PK, be able to extend once pkg/script/interpreter.go is created
	// todo() we must also have access to the previous tx output in order to verify the ScriptPubKey script
	txData := tx.AssembleForSigning()

	for _, vin := range tx.Vin {
		if vin.CanUnlockOutputWith(string(w.PublicKey)) {
			signature, err := w.signer.Sign(txData, w.PrivateKey)
			if err != nil {
				return nil, err
			}

			vin.ScriptSig = string(signature)
		}
	}

	return tx, nil
}

// generateInputs
func generateInputs(utxos []*kernel.UnspentOutput, targetAmount uint) ([]kernel.TxInput, uint, error) {
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

// generateOutputs
func generateOutputs(targetAmount, txFee, totalBalance uint, receiver, changeReceiver string) ([]kernel.TxOutput, error) {
	return []kernel.TxOutput{
		kernel.NewOutput(targetAmount, script.P2PK, receiver),
		kernel.NewOutput(totalBalance-txFee-targetAmount, script.P2PK, changeReceiver),
	}, nil
}

// broadcastTransaction
func broadcastTransaction(tx *kernel.Transaction) (*kernel.Transaction, error) {
	// todo() implement the logic to broadcast the transaction to the network
	return tx, nil
}
