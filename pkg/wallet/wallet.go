package wallet

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	rpnInter "chainnet/pkg/script/interpreter"
	"errors"
	"fmt"

	base58 "github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	version    []byte
	PrivateKey []byte
	PublicKey  []byte

	id []byte

	validator consensus.LightValidator
	signer    sign.Signature
	// hasher used for deriving wallet related values
	walletHasher hash.Hashing
	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	interpreter     *rpnInter.RPNInterpreter
}

func (w *Wallet) ID() string {
	return string(w.id)
}

func NewWallet(version []byte, validator consensus.LightValidator, signer sign.Signature, walletHasher hash.Hashing, consensusHasher hash.Hashing) (*Wallet, error) {
	publicKey, privateKey, err := signer.NewKeyPair()
	if err != nil {
		return nil, err
	}

	id, err := walletHasher.Hash(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not hash the public key: %w", err)
	}

	return &Wallet{
		version:         version,
		PrivateKey:      privateKey,
		PublicKey:       publicKey,
		validator:       validator,
		id:              id,
		signer:          signer,
		walletHasher:    walletHasher,
		consensusHasher: consensusHasher,
		interpreter:     rpnInter.NewScriptInterpreter(signer),
	}, nil
}

// GetAddress returns one wallet address
// todo() implement hierarchically deterministic HD wallet
func (w *Wallet) GetAddress() ([]byte, error) {
	// hash the public key
	pubKeyHash, err := w.walletHasher.Hash(w.PublicKey)
	if err != nil {
		return []byte{}, fmt.Errorf("could not hash the public key: %w", err)
	}

	// add the version to the hashed public key in order to hash again and obtain the checksum
	versionedPayload := append(w.version, pubKeyHash...) //nolint:gocritic // we need to append the version to the payload
	checksum, err := w.walletHasher.Hash(versionedPayload)
	if err != nil {
		return []byte{}, fmt.Errorf("could not hash the versioned payload: %w", err)
	}

	// return the base58 of the versioned payload and the checksum
	payload := append(versionedPayload, checksum...) //nolint:gocritic // we need to append the checksum to the payload
	return []byte(base58.Encode(payload)), nil
}

// SendTransaction creates a transaction and broadcasts it to the network
func (w *Wallet) SendTransaction(to string, targetAmount uint, txFee uint, utxos []*kernel.UnspentOutput) (*kernel.Transaction, error) {
	// create the inputs necessary for the transaction
	inputs, totalBalance, err := generateInputs(utxos, targetAmount+txFee)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// create the outputs necessary for the transaction
	outputs := generateOutputs(targetAmount, txFee, totalBalance, to, string(w.PublicKey))

	// generate transaction
	tx := kernel.NewTransaction(
		inputs,
		outputs,
	)

	// unlock the funds from the UTXOs
	tx, err = w.UnlockTxFunds(tx, utxos)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// generate tx hash
	txHash, err := util.CalculateTxHash(tx, w.consensusHasher)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	tx.SetID(txHash)
	// perform simple validations (light validator) before broadcasting the transaction
	if err = w.validator.ValidateTxLight(tx); err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error validating transaction: %w", err)
	}

	return broadcastTransaction(tx)
}

// UnlockTxFunds take a tx that is being built and unlocks the UTXOs from which the input funds are going to
// be used
func (w *Wallet) UnlockTxFunds(tx *kernel.Transaction, utxos []*kernel.UnspentOutput) (*kernel.Transaction, error) {
	// todo() for now, this only applies to P2PK, be able to extend once pkg/script/interpreter.go is created
	scripSigs := []string{}
	for _, vin := range tx.Vin {
		unlocked := false

		for _, utxo := range utxos {
			if utxo.EqualInput(vin) {
				// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
				scriptSig, err := w.interpreter.GenerateScriptSig(utxo.Output.ScriptPubKey, w.PrivateKey, w.PublicKey, tx)
				if err != nil {
					return &kernel.Transaction{}, fmt.Errorf("couldn't generate scriptSig for input with ID %s and index %d: %w", vin.Txid, vin.Vout, err)
				}

				scripSigs = append(scripSigs, scriptSig)

				unlocked = true
				continue
			}
		}

		// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
		if !unlocked {
			return &kernel.Transaction{}, fmt.Errorf("couldn't unlock funds for input with ID %s and index %d", vin.Txid, vin.Vout)
		}
	}

	for i := range len(tx.Vin) {
		tx.Vin[i].ScriptSig = scripSigs[i]
	}

	return tx, nil
}

// generateInputs set up the inputs for the transaction and returns the total balance of the UTXOs that are going to be
// spent in the transaction
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

// generateOutputs set up the outputs for the transaction
func generateOutputs(targetAmount, txFee, totalBalance uint, receiver, changeReceiver string) []kernel.TxOutput {
	change := totalBalance - txFee - targetAmount

	txOutput := []kernel.TxOutput{}
	txOutput = append(txOutput, kernel.NewOutput(targetAmount, script.P2PK, receiver))

	// add output corresponding to the spare change
	if change > 0 {
		txOutput = append(txOutput, kernel.NewOutput(totalBalance-txFee-targetAmount, script.P2PK, changeReceiver))
	}

	return txOutput
}

// broadcastTransaction sends the transaction to the network
func broadcastTransaction(tx *kernel.Transaction) (*kernel.Transaction, error) {
	// todo() implement the logic to broadcast the transaction to the network
	return tx, nil
}
