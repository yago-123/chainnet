package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/consensus/util"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/p2p"
	"github.com/yago-123/chainnet/pkg/script"
	rpnInter "github.com/yago-123/chainnet/pkg/script/interpreter"

	"github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	version    []byte
	PrivateKey []byte
	PublicKey  []byte

	validator consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer sign.Signature

	p2pActive    bool
	p2pNet       *p2p.WalletP2P
	p2pCtx       context.Context
	p2pCancelCtx context.CancelFunc

	encoder encoding.Encoding

	// hasher used for deriving wallet related values
	walletHasher hash.Hashing
	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	interpreter     *rpnInter.RPNInterpreter

	cfg *config.Config
}

func NewWallet(
	cfg *config.Config,
	version []byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	walletHasher hash.Hashing,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
) (*Wallet, error) {
	publicKey, privateKey, err := signer.NewKeyPair()
	if err != nil {
		return nil, err
	}

	return NewWalletWithKeys(
		cfg,
		version,
		validator,
		signer,
		walletHasher,
		consensusHasher,
		encoder,
		privateKey,
		publicKey,
	)
}

func NewWalletWithKeys(
	cfg *config.Config,
	version []byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	walletHasher hash.Hashing,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
	publicKey []byte,
) (*Wallet, error) {
	return &Wallet{
		cfg:             cfg,
		version:         version,
		PrivateKey:      privateKey,
		PublicKey:       publicKey,
		validator:       validator,
		signer:          signer,
		encoder:         encoder,
		walletHasher:    walletHasher,
		consensusHasher: consensusHasher,
		interpreter:     rpnInter.NewScriptInterpreter(signer),
	}, nil
}

func (w *Wallet) InitNetwork() (*p2p.WalletP2P, error) {
	var p2pNet *p2p.WalletP2P

	// check if the network has been initialized before
	if w.p2pActive {
		return nil, fmt.Errorf("p2p network is already active")
	}

	// create new P2P node
	w.p2pCtx, w.p2pCancelCtx = context.WithCancel(context.Background())
	p2pNet, err := p2p.NewWalletP2P(w.p2pCtx, w.cfg, w.encoder)
	if err != nil {
		return nil, fmt.Errorf("could not create wallet p2p network: %w", err)
	}

	// start the p2p node
	if err = p2pNet.Start(); err != nil {
		return nil, fmt.Errorf("error starting p2p node: %w", err)
	}

	w.p2pNet = p2pNet
	w.p2pActive = true

	if err = p2pNet.ConnectToSeeds(); err != nil {
		return nil, fmt.Errorf("error connecting to seeds: %w", err)
	}

	return p2pNet, nil
}

func (w *Wallet) GetP2PKAddress() []byte {
	return w.PublicKey
}

// GetAddress returns one wallet address
// todo() implement hierarchically deterministic HD wallet
func (w *Wallet) GetP2PKHAddress() ([]byte, error) {
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

func (w *Wallet) GetWalletUTXOS() ([]*kernel.UTXO, error) {
	utxos := []*kernel.UTXO{}

	address := w.GetP2PKAddress()
	p2pkUtxos, err := w.p2pNet.GetWalletUTXOS(address)
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("could not get wallet UTXOs for P2PK: %w", err)
	}

	address, err = w.GetP2PKHAddress()
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("could not get wallet address for P2PKH: %w", err)
	}
	p2pkhUtxos, err := w.p2pNet.GetWalletUTXOS(address)
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("could not get wallet UTXOs for P2PKH: %w", err)
	}

	// todo() add more types of addresses when are ready (multisig, etc)

	utxos = append(utxos, p2pkUtxos...)
	utxos = append(utxos, p2pkhUtxos...)

	return utxos, nil
}

// GenerateNewTransaction creates a transaction and broadcasts it to the network
func (w *Wallet) GenerateNewTransaction(to string, targetAmount uint, txFee uint, utxos []*kernel.UTXO) (*kernel.Transaction, error) {
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

	// assign the tx hash
	tx.SetID(txHash)

	// perform simple validations (light validator) before broadcasting the transaction
	if err = w.validator.ValidateTxLight(tx); err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error validating transaction: %w", err)
	}

	return tx, nil
}

// UnlockTxFunds take a tx that is being built and unlocks the UTXOs from which the input funds are going to
// be used
func (w *Wallet) UnlockTxFunds(tx *kernel.Transaction, utxos []*kernel.UTXO) (*kernel.Transaction, error) {
	// todo() for now, this only applies to P2PK, be able to extend once pkg/script/interpreter.go is created
	scripSigs := []string{}
	for _, vin := range tx.Vin {
		unlocked := false

		for _, utxo := range utxos {
			if utxo.EqualInput(vin) {
				// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
				scriptSig, err := w.interpreter.GenerateScriptSig(utxo.Output.ScriptPubKey, w.PublicKey, w.PrivateKey, tx)
				if err != nil {
					return &kernel.Transaction{}, fmt.Errorf("couldn't generate scriptSig for input with ID %x and index %d: %w", vin.Txid, vin.Vout, err)
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

// SendTransaction propagates a transaction to the network
func (w *Wallet) SendTransaction(ctx context.Context, tx *kernel.Transaction) error {
	if err := w.p2pNet.SendTransaction(ctx, *tx); err != nil {
		return fmt.Errorf("error sending transaction %x to the network: %w", tx.ID, err)
	}

	return nil
}

// generateInputs set up the inputs for the transaction and returns the total balance of the UTXOs that are going to be
// spent in the transaction
func generateInputs(utxos []*kernel.UTXO, targetAmount uint) ([]kernel.TxInput, uint, error) {
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
