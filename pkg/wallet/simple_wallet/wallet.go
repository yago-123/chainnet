package simplewallet

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcutil/base58"

	common "github.com/yago-123/chainnet/pkg/wallet"

	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"

	sdkv1beta "github.com/yago-123/chainnet-sdk-go/v1beta"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
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
	signer     sign.Signature
	nodeClient *sdkv1beta.Client
	encoder    encoding.Encoding

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
	nodeClient, err := sdkv1beta.NewClient(
		fmt.Sprintf("%s:%d", cfg.Wallet.ServerAddress, cfg.Wallet.ServerPort),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create wallet node client: %w", err)
	}

	return &Wallet{
		cfg:             cfg,
		version:         version,
		privateKey:      privateKey,
		publicKey:       publicKey,
		validator:       validator,
		signer:          signer,
		nodeClient:      nodeClient,
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

func (w *Wallet) GetWalletUTXOS() ([]sdkv1beta.UTXO, error) {
	addresses, err := w.GetAddresses()
	if err != nil {
		return []sdkv1beta.UTXO{}, fmt.Errorf("could not get wallet addresses: %w", err)
	}

	utxos := make([]sdkv1beta.UTXO, 0)
	for _, address := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), w.cfg.P2P.ConnTimeout)
		defer cancel()

		// retrieve UTXOs for each address
		utxo, errUtxos := w.nodeClient.GetAddressUTXOs(ctx, address)
		if errUtxos != nil {
			return []sdkv1beta.UTXO{}, fmt.Errorf("could not get wallet UTXOs for address %s: %w", base58.Encode(address), errUtxos)
		}

		utxos = append(utxos, utxo...)
	}

	return utxos, nil
}

// GetWalletTxs retrieves all the transactions that are related to the wallet
func (w *Wallet) GetWalletTxs() ([]*sdkv1beta.Transaction, error) {
	addresses, err := w.GetAddresses()
	if err != nil {
		return []*sdkv1beta.Transaction{}, fmt.Errorf("could not get wallet addresses: %w", err)
	}

	txs := make([]*sdkv1beta.Transaction, 0)
	for _, address := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), w.cfg.P2P.ConnTimeout)
		defer cancel()

		// retrieve txs for each address
		tx, errTxs := w.nodeClient.GetAddressTransactions(ctx, address)
		if errTxs != nil {
			return []*sdkv1beta.Transaction{}, fmt.Errorf("could not get wallet transactions for address %s: %w", base58.Encode(address), errTxs)
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
		active, errNet := w.nodeClient.AddressIsActive(ctx, address)
		if errNet != nil {
			return false, fmt.Errorf("could not check if address %s is active: %w", base58.Encode(address), errNet)
		}

		if active {
			return true, nil
		}
	}

	return false, nil
}

// GenerateNewTransaction creates a transaction using wallet-owned SDK types.
func (w *Wallet) GenerateNewTransaction(scriptType script.ScriptType, addresses []byte, targetAmount uint, txFee uint, utxos []sdkv1beta.UTXO) (*sdkv1beta.Transaction, error) {
	// create the inputs necessary for the transaction
	inputs, totalBalance, err := common.GenerateInputs(common.SDKUTXOsToKernel(utxos), targetAmount+txFee)
	if err != nil {
		return &sdkv1beta.Transaction{}, err
	}

	// create the outputs necessary for the transaction
	outputs, err := common.GenerateOutputs(scriptType, []uint{targetAmount}, [][]byte{addresses}, txFee, totalBalance, w.publicKey, w.version)
	if err != nil {
		return &sdkv1beta.Transaction{}, err
	}

	// generate transaction
	tx := kernel.NewTransaction(
		inputs,
		outputs,
	)

	// unlock the funds from the UTXOs
	sdkTx := common.KernelTransactionToSDK(*tx)
	sdkTxPtr, err := w.UnlockTxFunds(&sdkTx, utxos)
	if err != nil {
		return &sdkv1beta.Transaction{}, err
	}

	// generate tx hash
	txHash, err := util.CalculateTxHash(common.SDKTransactionToKernel(sdkTxPtr), w.consensusHasher)
	if err != nil {
		return &sdkv1beta.Transaction{}, err
	}

	// assign the tx hash
	sdkTxPtr.ID = txHash

	// perform simple validations (light validator) before broadcasting the transaction
	if err = w.validator.ValidateTxLight(common.SDKTransactionToKernel(sdkTxPtr)); err != nil {
		return &sdkv1beta.Transaction{}, fmt.Errorf("error validating transaction: %w", err)
	}

	return sdkTxPtr, nil
}

// UnlockTxFunds take a tx that is being built and unlocks the UTXOs from which the input funds are going to
// be used
func (w *Wallet) UnlockTxFunds(tx *sdkv1beta.Transaction, utxos []sdkv1beta.UTXO) (*sdkv1beta.Transaction, error) {
	// todo() for now, this only applies to P2PK, be able to extend once pkg/script/interpreter.go is created
	kernelTx := common.SDKTransactionToKernel(tx)
	kernelUtxos := common.SDKUTXOsToKernel(utxos)

	scriptSigs := []string{}
	for _, vin := range kernelTx.Vin {
		unlocked := false

		for _, utxo := range kernelUtxos {
			if utxo.EqualInput(vin) {
				// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
				scriptSig, err := w.interpreter.GenerateScriptSig(utxo.Output.ScriptPubKey, w.publicKey, w.privateKey, kernelTx)
				if err != nil {
					return &sdkv1beta.Transaction{}, fmt.Errorf("couldn't generate scriptSig for input with ID %x and index %d: %w", vin.Txid, vin.Vout, err)
				}

				scriptSigs = append(scriptSigs, scriptSig)

				unlocked = true
				continue
			}
		}

		// todo(): modify to allow multiple inputs with different scriptPubKeys owners (multiple wallets)
		if !unlocked {
			return &sdkv1beta.Transaction{}, fmt.Errorf("couldn't unlock funds for input with ID %x and index %d", vin.Txid, vin.Vout)
		}
	}

	for i := range len(kernelTx.Vin) {
		kernelTx.Vin[i].ScriptSig = scriptSigs[i]
	}

	sdkTx := common.KernelTransactionToSDK(*kernelTx)
	return &sdkTx, nil
}

// SendTransaction propagates a transaction to the network
func (w *Wallet) SendTransaction(ctx context.Context, tx sdkv1beta.Transaction) error {
	if err := w.nodeClient.SendTransaction(ctx, tx); err != nil {
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

func (w *Wallet) PrivateKey() []byte {
	return w.privateKey
}

func (w *Wallet) GetP2PKAddress() []byte {
	return w.publicKey
}

func (w *Wallet) GetP2PKHAddress() ([]byte, error) {
	return util_p2pkh.GenerateP2PKHAddrFromPubKey(w.publicKey, w.version)
}
