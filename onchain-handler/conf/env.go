package conf

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
	SSLMode    bool   `mapstructure:"DB_SSL_MODE"`
}

type PaymentGatewayConfiguration struct {
	InitWalletCount        uint   `mapstructure:"INIT_WALLET_COUNT"`
	ExpiredOrderTime       uint   `mapstructure:"EXPIRED_ORDER_TIME"`
	OrderCutoffTime        uint   `mapstructure:"ORDER_CUTOFF_TIME"`
	PaymentCovering        string `mapstructure:"PAYMENT_COVERING"`
	MasterWalletAddress    string `mapstructure:"MASTER_WALLET_ADDRESS"`
	WithdrawWorkerInterval string `mapstructure:"WITHDRAW_WORKER_INTERVAL"`
}

type BlockchainConfiguration struct {
	AvaxNetwork         AvaxNetworkConfiguration `mapstructure:",squash"`
	BscNetwork          BscNetworkConfiguration  `mapstructure:",squash"`
	GasBufferMultiplier string                   `mapstructure:"GAS_BUFFER_MULTIPLIER"`
}

type AvaxNetworkConfiguration struct {
	AvaxRpcUrls             string `mapstructure:"AVAX_RPC_URLS"`
	AvaxChainID             uint32 `mapstructure:"AVAX_CHAIN_ID"`
	AvaxStartBlockListener  uint64 `mapstructure:"AVAX_START_BLOCK_LISTENER"`
	AvaxUSDTContractAddress string `mapstructure:"AVAX_USDT_CONTRACT_ADDRESS"`
}

type BscNetworkConfiguration struct {
	BscRpcUrls             string `mapstructure:"BSC_RPC_URLS"`
	BscChainID             uint32 `mapstructure:"BSC_CHAIN_ID"`
	BscStartBlockListener  uint64 `mapstructure:"BSC_START_BLOCK_LISTENER"`
	BscUSDTContractAddress string `mapstructure:"BSC_USDT_CONTRACT_ADDRESS"`
}

type WalletConfiguration struct {
	Mnemonic   string `mapstructure:"MNEMONIC"`
	Passphrase string `mapstructure:"PASSPHRASE"`
	Salt       string `mapstructure:"SALT"`
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
	LogLevel       string                      `mapstructure:"LOG_LEVEL"`
}

var configuration Configuration

var defaultConfigurations = map[string]any{
	"REDIS_ADDRESS":              "localhost:6379",
	"REDIS_TTL":                  "60",
	"APP_PORT":                   "8080",
	"ENV_FILE":                   ".env",
	"ENV":                        "DEV",
	"LOG_LEVEL":                  "debug",
	"DB_USER":                    "",
	"DB_PASSWORD":                "",
	"DB_HOST":                    "",
	"DB_PORT":                    "",
	"DB_NAME":                    "",
	"DB_SSL_MODE":                false,
	"INIT_WALLET_COUNT":          10,
	"ORDER_CUTOFF_TIME":          1440,
	"EXPIRED_ORDER_TIME":         15,
	"PAYMENT_COVERING":           1,
	"GAS_BUFFER_MULTIPLIER":      2,
	"WITHDRAW_WORKER_INTERVAL":   "hourly",
	"MASTER_WALLET_ADDRESS":      "",
	"AVAX_RPC_URLS":              "",
	"AVAX_CHAIN_ID":              0,
	"AVAX_START_BLOCK_LISTENER":  0,
	"AVAX_USDT_CONTRACT_ADDRESS": "",
	"BSC_RPC_URLS":               "",
	"BSC_CHAIN_ID":               0,
	"BSC_START_BLOCK_LISTENER":   0,
	"BSC_USDT_CONTRACT_ADDRESS":  "",
	"MNEMONIC":                   "",
	"PASSPHRASE":                 "",
	"SALT":                       "",
}

// loadDefaultConfigs sets default values for critical configurations
func loadDefaultConfigs() {
	for configKey, configValue := range defaultConfigurations {
		viper.SetDefault(configKey, configValue)
	}
}

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	viper.SetConfigFile(envFile)
	viper.AutomaticEnv() // Automatically read from env variables

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults for critical configurations
	loadDefaultConfigs()

	// Attempt to read config file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file \"%s\" not found, falling back to environment variables", envFile)
	}

	// Unmarshal values into struct
	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatalf("Error unmarshalling configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
}

func GetRpcUrls(network constants.NetworkType) ([]string, error) {
	var rpcUrls string

	switch network {
	case constants.Bsc:
		rpcUrls = configuration.Blockchain.BscNetwork.BscRpcUrls
	case constants.AvaxCChain:
		rpcUrls = configuration.Blockchain.AvaxNetwork.AvaxRpcUrls
	default:
		return nil, fmt.Errorf("unsupported network type: %s", network)
	}

	if rpcUrls == "" {
		return nil, fmt.Errorf("no RPC URLs configured for network: %s", network)
	}

	// Split the RPC URLs by comma and trim spaces for each URL
	urls := strings.Split(rpcUrls, ",")
	for i, url := range urls {
		urls[i] = strings.TrimSpace(url)
	}

	return urls, nil
}

func GetConfiguration() *Configuration {
	return &configuration
}

func GetRedisConnectionURL() string {
	return configuration.Redis.RedisAddress
}

func (config *Configuration) GetExpiredOrderTime() time.Duration {
	return time.Duration(config.PaymentGateway.ExpiredOrderTime) * time.Minute
}

func (config *Configuration) GetOrderCutoffTime() time.Duration {
	return time.Duration(config.PaymentGateway.OrderCutoffTime) * time.Minute
}

func (config *Configuration) GetPaymentCovering() float64 {
	paymentCoveringStr := config.PaymentGateway.PaymentCovering
	if paymentCoveringStr == "" {
		log.Println("PaymentCovering is not set or is empty in the configuration")
		return 0.0
	}

	// Convert string to float64
	paymentCoveringFloat, err := strconv.ParseFloat(paymentCoveringStr, 64)
	if err != nil {
		log.Printf("Error parsing PaymentCovering as float64: %v. Using default value: 0", err)
		return 0.0
	}

	if paymentCoveringFloat < 0 {
		log.Printf("PaymentCovering must be greater than or equal 0. Using default value: 0")
		return 0.0
	}

	return paymentCoveringFloat
}

func (config *Configuration) GetGasBufferMultiplier() float64 {
	multiplierStr := config.Blockchain.GasBufferMultiplier
	if multiplierStr == "" {
		log.Println("GetGasBufferMultiplier is not set or is empty in the configuration")
		return 1.0
	}

	multiplier, err := strconv.ParseFloat(multiplierStr, 64)
	if err != nil {
		log.Printf("Invalid GetGasBufferMultiplier: %s. Using default value: 1. Error: %v", multiplierStr, err)
		return 1.0
	}

	return multiplier
}

func (config *Configuration) GetTokenSymbol(tokenAddress string) (string, error) {
	tokenSymbols := map[string]string{
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress: constants.USDT,
		config.Blockchain.BscNetwork.BscUSDTContractAddress:   constants.USDT,
	}

	if symbol, exists := tokenSymbols[tokenAddress]; exists {
		return symbol, nil
	}
	return "", fmt.Errorf("unknown token address: %s", tokenAddress)
}

func (config *Configuration) GetTokenAddress(symbol, network string) (string, error) {
	tokenAddresses := map[string]map[string]string{
		constants.AvaxCChain.String(): {
			constants.USDT: config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		},
		constants.Bsc.String(): {
			constants.USDT: config.Blockchain.BscNetwork.BscUSDTContractAddress,
		},
	}

	// Check if the network exists in the mapping
	if tokensForNetwork, exists := tokenAddresses[network]; exists {
		// Check if the symbol exists in the network's tokens
		if address, exists := tokensForNetwork[symbol]; exists {
			return address, nil
		}
		return "", fmt.Errorf("unknown token symbol for network %s: %s", network, symbol)
	}
	return "", fmt.Errorf("unsupported network: %s", network)
}
