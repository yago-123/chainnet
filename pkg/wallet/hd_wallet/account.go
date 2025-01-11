package hd_wallet

import (
	"context"
	"fmt"
	"sync"

	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/network"
	"github.com/yago-123/chainnet/pkg/script"
	rpnInter "github.com/yago-123/chainnet/pkg/script/interpreter"
	"github.com/yago-123/chainnet/pkg/util"
	common "github.com/yago-123/chainnet/pkg/wallet"

	"github.com/sirupsen/logrus"
	wallt "github.com/yago-123/chainnet/pkg/wallet/simple_wallet"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	cerror "github.com/yago-123/chainnet/pkg/errs"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
)

type Account struct {
	// derivedPubAccountKey public key derived from the master private key to be used for this account
	// in other words, level 3 of the HD wallet derivation path
	derivedPubAccountKey []byte
	// derivedChainAccountCode chain code derived from the original master private key to be used for this account
	derivedChainAccountCode []byte
	// accountID represents the number that corresponds to this account (constant for each account)
	accountID uint32

	externalWallets []*wallt.Wallet
	internalWallets []*wallt.Wallet

	// mu mutex used to synchronize the access to the account fields
	mu sync.Mutex

	// todo(): maybe encapsulate these fields in a struct?
	walletVersion byte
	validator     consensus.LightValidator
	// signer used for signing transactions and creating pub and private keys
	signer sign.Signature

	p2pNet network.WalletNetwork

	encoder encoding.Encoding

	// consensusHasher used for deriving blockchain related values (tx ID for example)
	consensusHasher hash.Hashing
	interpreter     *rpnInter.RPNInterpreter

	logger *logrus.Logger
	cfg    *config.Config
}

func NewHDAccount(
	cfg *config.Config,
	walletVersion byte,
	validator consensus.LightValidator,
	signer sign.Signature,
	consensusHasher hash.Hashing,
	encoder encoding.Encoding,
	derivedPubAccountKey []byte,
	derivedChainAccountCode []byte,
	accountNum uint32,
) (*Account, error) {
	p2pNet, err := network.NewWalletHTTPConn(cfg, encoder)
	if err != nil {
		return nil, fmt.Errorf("could not create wallet p2p network: %w", err)
	}

	return &Account{
		cfg:                     cfg,
		logger:                  cfg.Logger,
		walletVersion:           walletVersion,
		validator:               validator,
		signer:                  signer,
		p2pNet:                  p2pNet,
		consensusHasher:         consensusHasher,
		interpreter:             rpnInter.NewScriptInterpreter(signer),
		encoder:                 encoder,
		derivedPubAccountKey:    derivedPubAccountKey,
		derivedChainAccountCode: derivedChainAccountCode,
		accountID:               accountNum,
	}, nil
}

// Sync scans and updates the account by generating a series of externalWallets based on the default gap limit.
// It checks each wallet for funds by syncing with the network and looking for transactions. The gap limit
// defines the maximum number of consecutive empty externalWallets that can be generated before stopping the sync process.
// If a wallet contains transactions, it is considered active, and the process continues. If an account has
// no funds (empty wallet) for the specified gap limit, the syncing process halts, and the number of active externalWallets
// is recorded.
// Returns the number of active externalWallets found during the sync process and an error if any
func (hda *Account) Sync() (uint32, uint32, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	// helper function to sync wallets
	syncWallets := func(changeType changeType) ([]*wallt.Wallet, error) {
		wallets := []*wallt.Wallet{}
		gapCounter := 0

		for {
			wallet, err := hda.createWallet(changeType, uint32(len(wallets)))
			if err != nil {
				return []*wallt.Wallet{}, fmt.Errorf("error creating wallet: %w", err)
			}

			wallets = append(wallets, wallet)

			txs, err := wallet.GetWalletTxs()
			if err != nil {
				return []*wallt.Wallet{}, fmt.Errorf("error getting wallet transactions: %w", err)
			}

			if len(txs) == 0 {
				gapCounter++
			}

			if len(txs) > 0 {
				gapCounter = 0
			}

			if gapCounter >= AddressGapLimit {
				break
			}
		}

		// remove wallets that exceeded the gap limit
		activeWallets := wallets[:len(wallets)-AddressGapLimit]
		return activeWallets, nil
	}

	// sync external wallets
	externalWallets, err := syncWallets(ExternalChangeType)
	if err != nil {
		return 0, 0, fmt.Errorf("failed syncing external addresses: %w", err)
	}
	hda.externalWallets = externalWallets

	// sync internal wallets
	internalWallets, err := syncWallets(InternalChangeType)
	if err != nil {
		return 0, 0, fmt.Errorf("failed syncing internal addresses: %w", err)
	}
	hda.internalWallets = internalWallets

	return uint32(len(hda.externalWallets)), uint32(len(hda.internalWallets)), nil
}

func (hda *Account) GetAccountID() uint32 {
	return hda.accountID
}

// GetNewInternalWallet generates a new internal wallet
func (hda *Account) GetNewInternalWallet() (*wallt.Wallet, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	return hda.getNewWallet(InternalChangeType, &hda.internalWallets)
}

// GetNewExternalWallet generates a new external wallet
func (hda *Account) GetNewExternalWallet() (*wallt.Wallet, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	return hda.getNewWallet(ExternalChangeType, &hda.externalWallets)
}

func (hda *Account) GetExternalWalletIndex() uint32 {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	return uint32(len(hda.externalWallets))
}

func (hda *Account) GetInternalWalletIndex() uint32 {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	return uint32(len(hda.internalWallets))
}

// GetAccountUTXOs retrieves the UTXOs from both external and internal wallets. The resulting array is sorted by default
// with external UTXOs first
func (hda *Account) GetAccountUTXOs() ([]*kernel.UTXO, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	utxosCollection := []*kernel.UTXO{}
	wallets := append(hda.externalWallets, hda.internalWallets...)
	for _, wall := range wallets {
		utxos, err := wall.GetWalletUTXOS()
		if err != nil {
			return nil, err
		}
		utxosCollection = append(utxosCollection, utxos...)
	}

	return utxosCollection, nil
}

func (hda *Account) GenerateNewTransaction(scriptType script.ScriptType, addresses [][]byte, targetAmount []uint, txFee uint, utxos []*kernel.UTXO) (*kernel.Transaction, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	// create the inputs necessary for the transaction
	totalTargetAmount := uint(0)
	for _, amount := range targetAmount {
		totalTargetAmount += amount
	}

	inputs, totalBalance, err := common.GenerateInputs(utxos, totalTargetAmount+txFee)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// create the outputs necessary for the transaction
	changeWallet, err := hda.getNewWallet(InternalChangeType, &hda.internalWallets)
	if err != nil {
		return nil, err
	}

	outputs, err := common.GenerateOutputs(scriptType, targetAmount, addresses, txFee, totalBalance, changeWallet.PublicKey(), changeWallet.Version())
	if err != nil {
		return nil, err
	}

	// generate transaction
	tx := kernel.NewTransaction(
		inputs,
		outputs,
	)

	// unlock the funds from the UTXOs
	tx, err = hda.UnlockTxFunds(tx, utxos)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// generate tx hash
	txHash, err := util.CalculateTxHash(tx, hda.consensusHasher)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	// assign the tx hash
	tx.SetID(txHash)

	// perform simple validations (light validator) before broadcasting the transaction
	if err = hda.validator.ValidateTxLight(tx); err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error validating transaction: %w", err)
	}

	return tx, nil
}

func (hda *Account) UnlockTxFunds(tx *kernel.Transaction, utxos []*kernel.UTXO) (*kernel.Transaction, error) {
	// precompute wallets for lookup
	wallets := append(hda.externalWallets, hda.internalWallets...)

	// map UTXOs so that can be easily accessed for generating the scriptSigs for the inputs
	utxoMap := make(map[string]*kernel.UTXO)
	for _, utxo := range utxos {
		utxoMap[utxo.UniqueKey()] = utxo
	}

	scriptSigs := make([]string, len(tx.Vin))
	// iterate through each input and unlock the funds
	for i, vin := range tx.Vin {
		// retrieve the UTXO that will spend the input so that the scriptPubKey can be analyzed and unlocked
		utxo, found := utxoMap[vin.UniqueTxoKey()]
		if !found {
			return nil, fmt.Errorf("no matching UTXO found for input with ID %s and index %d", vin.Txid, vin.Vout)
		}

		// range over wallets to find the one that can unlock the funds
		// todo(): optimize this, maybe we can map the wallets with the public key
		unlocked := false
		for _, wallet := range wallets {
			if script.CanBeUnlockedWith(utxo.Output.ScriptPubKey, wallet.PublicKey(), wallet.Version()) {
				// generate the unlocking script
				// todo(1/2): remove the getter for PrivateKey, not correct to access it directly, there must be some entity
				// todo(2/2): or some logic rewriting
				scriptSig, err := hda.interpreter.GenerateScriptSig(utxo.Output.ScriptPubKey, wallet.PublicKey(), wallet.PrivateKey(), tx)
				if err != nil {
					return nil, fmt.Errorf("couldn't generate scriptSig for input with ID %x and index %d: %w", vin.Txid, vin.Vout, err)
				}

				// save the scriptSig and jump to the next input pending unlocking
				scriptSigs[i] = scriptSig
				unlocked = true
				break
			}
		}

		if !unlocked {
			return nil, fmt.Errorf("couldn't unlock funds for input with ID %s and index %d", vin.Txid, vin.Vout)
		}
	}

	// assign scriptSigs to transaction inputs
	for i := range tx.Vin {
		tx.Vin[i].ScriptSig = scriptSigs[i]
	}

	return tx, nil
}

// SendTransaction propagates a transaction to the network
func (hda *Account) SendTransaction(ctx context.Context, tx *kernel.Transaction) error {
	if err := hda.p2pNet.SendTransaction(ctx, *tx); err != nil {
		return fmt.Errorf("error sending transaction %x to the network: %w", tx.ID, err)
	}

	return nil
}

// GetBalance returns the total balance of the account by summing the amount of all UTXOs in the externalWallets
func (hda *Account) GetBalance() (uint, error) {
	hda.mu.Lock()
	defer hda.mu.Unlock()

	balance := uint(0)
	wallets := append(hda.externalWallets, hda.internalWallets...)
	for _, wallet := range wallets {
		utxos, err := wallet.GetWalletUTXOS()
		if err != nil {
			return 0, fmt.Errorf("error getting wallet UTXOs: %w", err)
		}

		for _, utxo := range utxos {
			balance += utxo.GetAmount()
		}
	}

	return balance, nil
}

func (hda *Account) ConsolidateChange() {

}

// getNewWallet generates a new wallet based on the specified change type
// Arguments:
// - changeType: represent the type of wallet that will be created (external or internal)
// - walletCollection: represent the collection in which the new wallet will be persisted (this can be internalWallets
// or externalWallets collection
// Returns:
// - the wallet generated, so it can be used
// - an error if any
func (hda *Account) getNewWallet(changeType changeType, walletCollection *[]*wallt.Wallet) (*wallt.Wallet, error) {
	wallet, err := hda.createWallet(changeType, uint32(len(*walletCollection)))
	if err != nil {
		return nil, fmt.Errorf("error creating new wallet: %w", err)
	}

	// append the new wallet to the appropriate wallet slice
	*walletCollection = append(*walletCollection, wallet)
	return wallet, nil
}

// createWallet generates a new wallet by deriving by default the external and the wallet number selected. Although this
// method is called createWallet, it should be called createAddress according to BIP-44, but given that all the code
// is already written for a simple wallet, it's better to keep it this way for now and reuse the code related to wallet.
// Also have the advantage that it will isolate the network traces. This method does not persist the wallet in the
// Account object itself, it's the responsibility of the caller to do so if needed
func (hda *Account) createWallet(change changeType, walletNum uint32) (*wallt.Wallet, error) {
	var err error
	var derivedPrivateKey []byte

	// derive the child key step by step, following the BIP44 path purpose' / coin type' / account' / change / index
	// where ' denotes hardened keys. The first three levels require hardened key by BIP44, in this case we are deriving
	// the account, so we only need the first three levels
	indexes := []uint32{
		uint32(change), // given that the wallet does not have funds yet, the change type is external by default
		walletNum,
	}

	derivedPublicKey, derivedChainCode := hda.derivedPubAccountKey, hda.derivedChainAccountCode

	// for each index in the derivation path, derive the child key
	for _, idx := range indexes {
		// the derivedKey field is a public key, but the return value will be a private and chain code key
		derivedPrivateKey, derivedChainCode, err = DeriveChildStepNonHardened(derivedPublicKey, derivedChainCode, idx)
		if err != nil {
			return nil, fmt.Errorf("error deriving child key: %w", err)
		}

		// extract public key from derived key to be used for the subsequent non-hardened indexes OR the wallet creation
		derivedPublicKey, err = util_crypto.DeriveECDSAPubFromPrivateDERBytes(derivedPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", cerror.ErrCryptoPublicKeyDerivation, err)
		}
	}

	// generate new wallet with the derived keys
	wallet, err := wallt.NewWalletWithKeys(
		hda.cfg,
		hda.walletVersion,
		hda.validator,
		hda.signer,
		hda.consensusHasher,
		hda.encoder,
		derivedPrivateKey,
		derivedPublicKey,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up wallet: %w", err)
	}

	return wallet, nil
}
