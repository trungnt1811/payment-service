package conf

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
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
	CacheType      string                      `mapstructure:"CACHE_TYPE"`
	WorkerEnabled  bool                        `mapstructure:"WORKER_ENABLED"`
}

var configuration Configuration

var defaultConfigurations = map[string]any{
	"REDIS_ADDRESS":              "localhost:6379",
	"REDIS_TTL":                  "60m",
	"APP_PORT":                   "8080",
	"APP_NAME":                   "onchain-handler",
	"ENV_FILE":                   ".env",
	"ENV":                        "DEV",
	"LOG_LEVEL":                  "debug",
	"CACHE_TYPE":                 "in-memory",
	"WORKER_ENABLED":             true,
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
	// Set environment variable for .env file location
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env" // Default to .env if ENV_FILE is not set
	}

	// Set Viper to look for the config file
	viper.SetConfigFile(envFile)
	viper.SetConfigType("env")                             // Explicitly tell Viper it's an .env file
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores

	// Set default values before reading config
	loadDefaultConfigs()

	// Attempt to read the .env file
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Loaded configuration from file: %s", envFile)
	} else {
		viper.AutomaticEnv() // Enable reading from environment variables
		log.Printf("Config file \"%s\" not found or unreadable, falling back to environment variables", envFile)
	}

	// Unmarshal values into the global `configuration` struct
	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatalf("Error unmarshalling configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
}
