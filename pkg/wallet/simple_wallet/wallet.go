package simplewallet

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcutil/base58"

	common "github.com/yago-123/chainnet/pkg/wallet"

	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/network"
	"github.com/yago-123/chainnet/pkg/script"
	rpnInter "github.com/yago-123/chainnet/pkg/script/interpreter"
	"github.com/yago-123/chainnet/pkg/util"
)

type Wallet struct {
	version    byte
	privateKey []byte
	publicKey  []byte

	validator consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer sign.Signature
	// p2pNet used for broadcasting transactions
	p2pNet  network.WalletNetwork
	encoder encoding.Encoding

	// hasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	interpreter     *rpnInter.RPNInterpreter

	cfg *config.Config
}

func NewWallet(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
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
		consensusHasher,
		encoder,
		privateKey,
		publicKey,
	)
}

func NewWalletWithKeys(
	cfg *config.Config,
	version byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	privateKey []byte,
	publicKey []byte,
) (*Wallet, error) {
	p2pNet, err := network.NewWalletHTTPConn(cfg, encoder)
	if err != nil {
		return nil, fmt.Errorf("could not create wallet p2p network: %w", err)
	}

	return &Wallet{
		cfg:             cfg,
		version:         version,
		privateKey:      privateKey,
		publicKey:       publicKey,
		validator:       validator,
		signer:          signer,
		p2pNet:          p2pNet,
		encoder:         encoder,
		consensusHasher: consensusHasher,
		interpreter:     rpnInter.NewScriptInterpreter(signer),
	}, nil
}

func (w *Wallet) GetAddresses() ([][]byte, error) {
	addresses := make([][]byte, 0)

	// retrieve P2PK address
	addresses = append(addresses, w.GetP2PKAddress())

	// retrieve P2PKH address
	address, err := w.GetP2PKHAddress()
	if err != nil {
		return [][]byte{}, fmt.Errorf("could not get wallet address for P2PKH: %w", err)
	}

	addresses = append(addresses, address)

	// validate that are between the allowed ranges
	for _, addr := range addresses {
		if !util.IsValidAddress(addr) {
			return [][]byte{}, fmt.Errorf("invalid address format for address %x", addr)
		}
	}

	// todo() add more types of addresses when are ready (multisig, etc)

	return addresses, nil
}

func (w *Wallet) GetWalletUTXOS() ([]*kernel.UTXO, error) {
	addresses, err := w.GetAddresses()
	if err != nil {
		return []*kernel.UTXO{}, fmt.Errorf("could not get wallet addresses: %w", err)
	}

	utxos := make([]*kernel.UTXO, 0)
	for _, address := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), w.cfg.P2P.ConnTimeout)
		defer cancel()

		// retrieve UTXOs for each address
		utxo, errUtxos := w.p2pNet.GetWalletUTXOS(ctx, address)
		if errUtxos != nil {
			return []*kernel.UTXO{}, fmt.Errorf("could not get wallet UTXOs for address %s: %w", base58.Encode(address), errUtxos)
		}

		utxos = append(utxos, utxo...)
	}

	return utxos, nil
}

// GetWalletTxs retrieves all the transactions that are related to the wallet
func (w *Wallet) GetWalletTxs() ([]*kernel.Transaction, error) {
	addresses, err := w.GetAddresses()
	if err != nil {
		return []*kernel.Transaction{}, fmt.Errorf("could not get wallet addresses: %w", err)
	}

	txs := make([]*kernel.Transaction, 0)
	for _, address := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), w.cfg.P2P.ConnTimeout)
		defer cancel()

		// retrieve txs for each address
		tx, errUtxos := w.p2pNet.GetWalletTxs(ctx, address)
		if errUtxos != nil {
			return []*kernel.Transaction{}, fmt.Errorf("could not get wallet UTXOs for address %s: %w", base58.Encode(address), errUtxos)
		}

		txs = append(txs, tx...)
	}

	return txs, nil
}

// CheckAddressIsActive checks if there has been any transaction related to any of the addresses of the wallet
func (w *Wallet) CheckIfWalletIsActive() (bool, error) {
	addresses, err := w.GetAddresses()
	if err != nil {
		return false, fmt.Errorf("could not get wallet addresses: %w", err)
	}

	for _, address := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), w.cfg.P2P.ConnTimeout)
		defer cancel()

		// check if address is active
		active, errNet := w.p2pNet.AddressIsActive(ctx, address)
		if errNet != nil {
			return false, fmt.Errorf("could not check if address %s is active: %w", base58.Encode(address), errNet)
		}

		if active {
			return true, nil
		}
	}

	return false, nil
}

// GenerateNewTransaction creates a transaction and broadcasts it to the network
func (w *Wallet) GenerateNewTransaction(scriptType script.ScriptType, addresses []byte, targetAmount uint, txFee uint, utxos []*kernel.UTXO) (*kernel.Transaction, error) {
	// create the inputs necessary for the transaction
	inputs, totalBalance, err := common.GenerateInputs(utxos, targetAmount+txFee)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// create the outputs necessary for the transaction
	outputs, err := common.GenerateOutputs(scriptType, []uint{targetAmount}, [][]byte{addresses}, txFee, totalBalance, w.publicKey, w.version)
	if err != nil {
		return &kernel.Transaction{}, err
	}

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
	scriptSigs := []string{}
	for _, vin := range tx.Vin {
		unlocked := false

		for _, utxo := range utxos {
			if utxo.EqualInput(vin) {
				// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
				scriptSig, err := w.interpreter.GenerateScriptSig(utxo.Output.ScriptPubKey, w.publicKey, w.privateKey, tx)
				if err != nil {
					return &kernel.Transaction{}, fmt.Errorf("couldn't generate scriptSig for input with ID %x and index %d: %w", vin.Txid, vin.Vout, err)
				}

				scriptSigs = append(scriptSigs, scriptSig)

				unlocked = true
				continue
			}
		}

		// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
		if !unlocked {
			return &kernel.Transaction{}, fmt.Errorf("couldn't unlock funds for input with ID %x and index %d", vin.Txid, vin.Vout)
		}
	}

	for i := range len(tx.Vin) {
		tx.Vin[i].ScriptSig = scriptSigs[i]
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

func (w *Wallet) Version() byte {
	return w.version
}

func (w *Wallet) PublicKey() []byte {
	return w.publicKey
}

// todo(): REMOVE THIS, THIS FEELS SO WRONG, NEED TO RETHINK THE RELATION ACCOUNT / WALLET / HD WALLET
func (w *Wallet) PrivateKey() []byte {
	return w.privateKey
}

func (w *Wallet) GetP2PKAddress() []byte {
	return w.publicKey
}

func (w *Wallet) GetP2PKHAddress() ([]byte, error) {
	return util_p2pkh.GenerateP2PKHAddrFromPubKey(w.publicKey, w.version)
}
