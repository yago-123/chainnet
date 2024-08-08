package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/storage"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sirupsen/logrus"
)

const (
	MiningInterval = 1 * time.Minute
)

var rootCmd = &cobra.Command{
	Use:   "node",
	Short: "Chainnet node app",
}

var cfg *config.Config
var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	// initialize config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// add general config flags
	config.AddConfigFlags(rootCmd)
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("unable to find config file: %s\n", err)

		fmt.Println("relying in default config file...")
		cfg = config.NewConfig()
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// Initialize configuration before executing commands
	initConfig()

	fmt.Printf("Loaded Configuration: %+v\n", cfg)

	// Execute the root command
	Execute()

	cfg.Logger.SetLevel(logrus.DebugLevel)

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB("bin/miner-storage", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create new observer
	subjectObserver := observer.NewSubjectObserver()

	// create new chain
	chain, err := blockchain.NewBlockchain(
		cfg,
		boltdb,
		hash.GetHasher(consensusHasherType),
		validator.NewHeavyValidator(
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb),
			crypto.NewHashedSignature(
				sign.NewECDSASignature(), hash.NewSHA256(),
			),
			hash.GetHasher(consensusHasherType),
		),
		subjectObserver,
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating blockchain: %s", err)
	}

	subjectObserver.Register(boltdb)

	chain.Sync() //?

	time.Sleep(MiningInterval)
}
