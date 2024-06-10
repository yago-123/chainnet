package cli

import (
	"chainnet/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger := logrus.New()

	// Load configuration values from viper
	difficultyPoW := viper.GetUint("DIFFICULTY_POW")
	maxNoncePoW := viper.GetUint("MAX_NONCE_POW")
	baseURL := viper.GetString("BASE_URL")

	cfg := config.NewConfig(logger, difficultyPoW, maxNoncePoW, baseURL)
	Execute(cfg)
}
