package conf

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/genefriendway/onchain-handler/constants"
)

type RedisConfiguration struct {
	RedisAddress string `mapstructure:"REDIS_ADDRESS"`
	RedisTtl     string `mapstructure:"REDIS_TTL"`
}

type DatabaseConfiguration struct {
	DbUser     string `mapstructure:"DB_USER"`
	DbPassword string `mapstructure:"DB_PASSWORD"`
	DbHost     string `mapstructure:"DB_HOST"`
	DbPort     string `mapstructure:"DB_PORT"`
	DbName     string `mapstructure:"DB_NAME"`
}

type PaymentGatewayConfiguration struct {
	InitWalletCount        uint   `mapstructure:"INIT_WALLET_COUNT"`
	ExpiredOrderTime       uint   `mapstructure:"EXPIRED_ORDER_TIME"`
	PaymentCovering        string `mapstructure:"PAYMENT_COVERING"`
	MasterWalletAddress    string `mapstructure:"MASTER_WALLET_ADDRESS"`
	PrivateKeyMasterWallet string `mapstructure:"PRIVATE_KEY_MASTER_WALLET"`
}

type BlockchainConfiguration struct {
	AvaxNetwork      AvaxNetworkConfiguration      `mapstructure:",squash"`
	BscNetwork       BscNetworkConfiguration       `mapstructure:",squash"`
	LPTreasuryPool   LPTreasuryPoolConfiguration   `mapstructure:",squash"`
	USDTTreasuryPool USDTTreasuryPoolConfiguration `mapstructure:",squash"`
	LPRevenuePool    LPRevenuePoolConfiguration    `mapstructure:",squash"`
}

type AvaxNetworkConfiguration struct {
	AvaxRpcUrl                    string `mapstructure:"AVAX_RPC_URL"`
	AvaxChainID                   uint32 `mapstructure:"AVAX_CHAIN_ID"`
	AvaxStartBlockListener        uint64 `mapstructure:"AVAX_START_BLOCK_LISTENER"`
	AvaxUSDTContractAddress       string `mapstructure:"AVAX_USDT_CONTRACT_ADDRESS"`
	AvaxLifePointContractAddress  string `mapstructure:"AVAX_LIFE_POINT_CONTRACT_ADDRESS"`
	AvaxBulkSenderContractAddress string `mapstructure:"AVAX_BULK_SENDER_CONTRACT_ADDRESS"`
}

type BscNetworkConfiguration struct {
	BscRpcUrl                    string `mapstructure:"BSC_RPC_URL"`
	BscChainID                   uint32 `mapstructure:"BSC_CHAIN_ID"`
	BscStartBlockListener        uint64 `mapstructure:"BSC_START_BLOCK_LISTENER"`
	BscUSDTContractAddress       string `mapstructure:"BSC_USDT_CONTRACT_ADDRESS"`
	BscBulkSenderContractAddress string `mapstructure:"BSC_BULK_SENDER_CONTRACT_ADDRESS"`
}

type WalletConfiguration struct {
	Mnemonic   string `mapstructure:"MNEMONIC"`
	Passphrase string `mapstructure:"PASSPHRASE"`
	Salt       string `mapstructure:"SALT"`
}

type LPTreasuryPoolConfiguration struct {
	LPTreasuryAddress    string `mapstructure:"LP_TREASURY_ADDRESS"`
	PrivateKeyLPTreasury string `mapstructure:"PRIVATE_KEY_LP_TREASURY"`
}

type USDTTreasuryPoolConfiguration struct {
	USDTTreasuryAddress    string `mapstructure:"USDT_TREASURY_ADDRESS"`
	PrivateKeyUSDTTreasury string `mapstructure:"PRIVATE_KEY_USDT_TREASURY"`
}

type LPRevenuePoolConfiguration struct {
	LPRevenueAddress    string `mapstructure:"LP_REVENUE_ADDRESS"`
	PrivateKeyLPRevenue string `mapstructure:"PRIVATE_KEY_LP_REVENUE"`
}

type Configuration struct {
	Database       DatabaseConfiguration       `mapstructure:",squash"`
	Redis          RedisConfiguration          `mapstructure:",squash"`
	Blockchain     BlockchainConfiguration     `mapstructure:",squash"`
	PaymentGateway PaymentGatewayConfiguration `mapstructure:",squash"`
	Wallet         WalletConfiguration         `mapstructure:",squash"`
	AppName        string                      `mapstructure:"APP_NAME"`
	AppPort        uint32                      `mapstructure:"APP_PORT"`
	Env            string                      `mapstructure:"ENV"`
}

var configuration Configuration

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	viper.SetConfigFile("./.env")
	viper.AutomaticEnv()

	// Set defaults for critical configurations
	viper.SetDefault("APP_PORT", 8080)
	viper.SetDefault("EXPIRED_ORDER_TIME", 15)

	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile(fmt.Sprintf("../%s", envFile))
		if err := viper.ReadInConfig(); err != nil {
			log.Logger.Printf("Error reading config file \"%s\", %v", envFile, err)
		}
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading config file %v", err)
	}

	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatal().Err(err).Msgf("Error unmarshalling configuration %v", err)
	}

	// Override configuration with environment variables
	configuration.PaymentGateway.PrivateKeyMasterWallet = os.Getenv("PRIVATE_KEY_MASTER_WALLET")
	configuration.Blockchain.LPTreasuryPool.PrivateKeyLPTreasury = os.Getenv("PRIVATE_KEY_LP_TREASURY")
	configuration.Blockchain.USDTTreasuryPool.PrivateKeyUSDTTreasury = os.Getenv("PRIVATE_KEY_USDT_TREASURY")
	configuration.Blockchain.LPRevenuePool.PrivateKeyLPRevenue = os.Getenv("PRIVATE_KEY_LP_REVENUE")
	configuration.Wallet.Mnemonic = os.Getenv("MNEMONIC")
	configuration.Wallet.Passphrase = os.Getenv("PASSPHRASE")
	configuration.Wallet.Salt = os.Getenv("SALT")
	log.Info().Msg("Configuration loaded successfully")
}

func GetConfiguration() *Configuration {
	return &configuration
}

func GetRedisConnectionURL() string {
	return configuration.Redis.RedisAddress
}

// Helper function to get pools' private key by pools' addresses
func (config Configuration) GetPoolPrivateKey(poolAddress string) (string, error) {
	switch poolAddress {
	// LP Treasury pool
	case config.Blockchain.LPTreasuryPool.LPTreasuryAddress:
		return config.Blockchain.LPTreasuryPool.PrivateKeyLPTreasury, nil
	// LP Revenue pool
	case config.Blockchain.LPRevenuePool.LPRevenueAddress:
		return config.Blockchain.LPRevenuePool.PrivateKeyLPRevenue, nil
	// USDT Treasury pool
	case config.Blockchain.USDTTreasuryPool.USDTTreasuryAddress:
		return config.Blockchain.USDTTreasuryPool.PrivateKeyUSDTTreasury, nil
	default:
		return "", fmt.Errorf("failed to get private key for pool address: %s", poolAddress)
	}
}

func (config Configuration) GetPoolAddress(poolName string) (string, error) {
	switch poolName {
	case constants.LPRevenue:
		return config.Blockchain.LPRevenuePool.LPRevenueAddress, nil
	case constants.LPTreasury:
		return config.Blockchain.LPTreasuryPool.LPTreasuryAddress, nil
	case constants.USDTTreasury:
		return config.Blockchain.USDTTreasuryPool.USDTTreasuryAddress, nil
	default:
		return "", fmt.Errorf("unrecognized pool name: %s", poolName)
	}
}

func (config Configuration) GetExpiredOrderTime() time.Duration {
	return time.Duration(config.PaymentGateway.ExpiredOrderTime) * time.Minute
}

func (config Configuration) GetPaymentCovering() float64 {
	paymentCoveringStr := config.PaymentGateway.PaymentCovering
	if paymentCoveringStr == "" {
		log.Error().Msg("PaymentCovering is not set or is empty in the configuration")
		return 0.0
	}

	// Convert string to float64
	paymentCoveringFloat, err := strconv.ParseFloat(paymentCoveringStr, 64)
	if err != nil {
		log.Error().Err(err).Str("PaymentCovering", paymentCoveringStr).Msg("Error parsing PaymentCovering as float64")
		return 0.0
	}

	if paymentCoveringFloat <= 0 {
		log.Error().Float64("PaymentCovering", paymentCoveringFloat).Msg("PaymentCovering must be greater than 0")
		return 0.0
	}

	return paymentCoveringFloat
}

func (config Configuration) GetTokenSymbol(tokenAddress string) (string, error) {
	tokenSymbols := map[string]string{
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress:      constants.USDT,
		config.Blockchain.BscNetwork.BscUSDTContractAddress:        constants.USDT,
		config.Blockchain.AvaxNetwork.AvaxLifePointContractAddress: constants.LP,
	}

	if symbol, exists := tokenSymbols[tokenAddress]; exists {
		return symbol, nil
	}
	return "", fmt.Errorf("unknown token address: %s", tokenAddress)
}
