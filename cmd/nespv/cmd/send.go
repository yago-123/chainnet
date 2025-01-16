package cmd

import (
	"context"

	"github.com/yago-123/chainnet/pkg/kernel"

	cerror "github.com/yago-123/chainnet/pkg/errs"
	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"

	"github.com/btcsuite/btcutil/base58"
	"github.com/yago-123/chainnet/pkg/script"

	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/encoding"
	wallt "github.com/yago-123/chainnet/pkg/wallet/simple_wallet"
)

const (
	FlagPayType   = "pay-type"
	FlagAddress   = "address"
	FlagAmount    = "amount"
	FlagFee       = "fee"
	FlagPrivKey   = "priv-key"
	FlagWalletKey = "wallet-key-path"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Long:  `Send transactions from wallets.`,
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = config.InitConfig(cmd)

		scriptTypeStr, _ := cmd.Flags().GetString(FlagPayType)
		address, _ := cmd.Flags().GetString(FlagAddress)
		amount, _ := cmd.Flags().GetFloat64(FlagAmount)
		fee, _ := cmd.Flags().GetFloat64(FlagFee)
		privKeyCont, _ := cmd.Flags().GetString(FlagPrivKey)
		privKeyPath, _ := cmd.Flags().GetString(FlagWalletKey)

		// check if only one private key is provided
		if (privKeyCont == "") == (privKeyPath == "") {
			logger.Fatalf("specify one argument containing the private key: --priv-key or --wallet-key-path")
		}

		var err error
		var privKey, pubKey []byte
		var payType script.ScriptType

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

		if scriptTypeStr != "" {
			payType = script.ReturnScriptTypeFromStringType(scriptTypeStr)
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

		utxos, err := wallet.GetWalletUTXOS()
		if err != nil {
			logger.Fatalf("error getting wallet UTXOS: %v", err)
		}

		tx, err := wallet.GenerateNewTransaction(payType, base58.Decode(address), kernel.ConvertFromCoinsToChannoshis(amount), kernel.ConvertFromCoinsToChannoshis(fee), utxos)
		if err != nil {
			logger.Fatalf("error generating transaction: %v", err)
		}

		context, cancel := context.WithTimeout(context.Background(), cfg.P2P.ConnTimeout)
		defer cancel()

		err = wallet.SendTransaction(context, tx)
		if err != nil {
			logger.Fatalf("error sending transaction: %v", err)
		}

		logger.Infof("Sent transaction: %s", tx.String())
	},
}

func init() {
	// main command
	config.AddConfigFlags(sendCmd)
	rootCmd.AddCommand(sendCmd)

	// sub commands
	sendCmd.Flags().String(FlagPayType, "P2PK", "Type of address to send coins to")
	sendCmd.Flags().String(FlagAddress, "", "Destination address to send coins")
	sendCmd.Flags().Float64(FlagAmount, 0.0, "Amount of coins to send")
	sendCmd.Flags().Float64(FlagFee, 0.0, "Amount of fee to send")
	sendCmd.Flags().String(FlagPrivKey, "", "Private key")

	// todo(): reestructure this duplication
	sendCmd.Flags().String(FlagWalletKey, "", "Path to private key")

	// required flags
	_ = sendCmd.MarkFlagRequired(FlagAddress)
	_ = sendCmd.MarkFlagRequired(FlagAmount)
}
