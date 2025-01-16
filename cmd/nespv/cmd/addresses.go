package cmd

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/encoding"
	cerror "github.com/yago-123/chainnet/pkg/errs"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"
	wallt "github.com/yago-123/chainnet/pkg/wallet/simple_wallet"
)

var addressesCmd = &cobra.Command{
	Use:   "addresses",
	Short: "Addresses wallet",
	Long:  `Retrieve addresses from wallet.`,
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = config.InitConfig(cmd)

		privKeyCont, _ := cmd.Flags().GetString(FlagPrivKey)
		privKeyPath, _ := cmd.Flags().GetString(FlagWalletKey)

		// check if only one private key is provided
		if (privKeyCont == "") == (privKeyPath == "") {
			logger.Fatalf("specify one argument containing the private key: --priv-key or --wallet-key-path")
		}

		var err error
		var privKey, pubKey []byte

		// process key from path or from content
		if privKeyCont != "" {
			// todo(): this is encoded somehow?
			privKey = base58.Decode(privKeyCont)
		}

		if privKeyPath != "" {
			privKey, err = util_crypto.ReadECDSAPemToPrivateKeyDerBytes(privKeyPath)
			if err != nil {
				logger.Fatalf("error reading private key: %v", err)
			}
		}

		// derive public key from private key
		pubKey, err = util_crypto.DeriveECDSAPubFromPrivateDERBytes(privKey)
		if err != nil {
			logger.Fatalf("%v: %v", cerror.ErrCryptoPublicKeyDerivation, err)
		}

		// create wallet
		wallet, err := wallt.NewWalletWithKeys(
			cfg,
			1,
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			consensusSigner,
			hash.GetHasher(consensusHasherType),
			encoding.NewProtobufEncoder(),
			privKey,
			pubKey,
		)
		if err != nil {
			logger.Fatalf("error setting up wallet: %v", err)
		}

		logger.Infof("P2PK addr: %s", base58.Encode(wallet.GetP2PKAddress()))

		p2pkhAddr, err := wallet.GetP2PKHAddress()
		if err != nil {
			logger.Fatalf("error getting P2PKH address: %v", err)
		}
		logger.Infof("P2PKH addr: %s", base58.Encode(p2pkhAddr))

		pubKeyHashedAddr, version, err := util_p2pkh.ExtractPubKeyHashedFromP2PKHAddr(p2pkhAddr)
		if err != nil {
			logger.Fatalf("error extracting pub key hash from P2PKH address: %v", err)
		}
		logger.Infof("Hashed-only P2PKH address %s, version: %d", base58.Encode(pubKeyHashedAddr), version)
	},
}

func init() {
	// main command
	config.AddConfigFlags(addressesCmd)
	rootCmd.AddCommand(addressesCmd)

	// sub commands
	addressesCmd.Flags().String(FlagPrivKey, "", "Private key")

	// todo(): reestructure this duplication
	addressesCmd.Flags().String(FlagWalletKey, "", "Path to private key")

	// required flags
	_ = addressesCmd.MarkFlagRequired(FlagAddress)
	_ = addressesCmd.MarkFlagRequired(FlagAmount)
}
