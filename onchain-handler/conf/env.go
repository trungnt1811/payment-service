package conf

import (
	"fmt"
	"os"

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

type BlockchainConfiguration struct {
	RpcUrl             string                        `mapstructure:"RPC_URL"`
	ChainID            uint32                        `mapstructure:"CHAIN_ID"`
	StartBlockListener uint64                        `mapstructure:"START_BLOCK_LISTENER"`
	SmartContract      SmartContractConfiguration    `mapstructure:",squash"`
	LPTreasuryPool     LPTreasuryPoolConfiguration   `mapstructure:",squash"`
	USDTTreasuryPool   USDTTreasuryPoolConfiguration `mapstructure:",squash"`
	LPCommunityPool    LPCommunityPoolConfiguration  `mapstructure:",squash"`
	LPRevenuePool      LPRevenuePoolConfiguration    `mapstructure:",squash"`
	LPStakingPool      LPStakingPoolConfiguration    `mapstructure:",squash"`
}

type LPTreasuryPoolConfiguration struct {
	LPTreasuryAddress    string `mapstructure:"LP_TREASURY_ADDRESS"`
	PrivateKeyLPTreasury string `mapstructure:"PRIVATE_KEY_LP_TREASURY"`
}

type USDTTreasuryPoolConfiguration struct {
	USDTTreasuryAddress    string `mapstructure:"USDT_TREASURY_ADDRESS"`
	PrivateKeyUSDTTreasury string `mapstructure:"PRIVATE_KEY_USDT_TREASURY"`
}

type LPCommunityPoolConfiguration struct {
	LPCommunityAddress    string `mapstructure:"LP_COMMUNITY_ADDRESS"`
	PrivateKeyLPCommunity string `mapstructure:"PRIVATE_KEY_LP_COMMUNITY"`
}

type LPRevenuePoolConfiguration struct {
	LPRevenueAddress    string `mapstructure:"LP_REVENUE_ADDRESS"`
	PrivateKeyLPRevenue string `mapstructure:"PRIVATE_KEY_LP_REVENUE"`
}

type LPStakingPoolConfiguration struct {
	LPStakingAddress    string `mapstructure:"LP_STAKING_ADDRESS"`
	PrivateKeyLPStaking string `mapstructure:"PRIVATE_KEY_LP_STAKING"`
}

type SmartContractConfiguration struct {
	LifePointContractAddress  string `mapstructure:"LIFE_POINT_CONTRACT_ADDRESS"`
	USDTContractAddress       string `mapstructure:"USDT_CONTRACT_ADDRESS"`
	BulkSenderContractAddress string `mapstructure:"BULK_SENDER_CONTRACT_ADDRESS"`
}

type Configuration struct {
	Database   DatabaseConfiguration   `mapstructure:",squash"`
	Redis      RedisConfiguration      `mapstructure:",squash"`
	Blockchain BlockchainConfiguration `mapstructure:",squash"`
	AppName    string                  `mapstructure:"APP_NAME"`
	AppPort    uint32                  `mapstructure:"APP_PORT"`
	Env        string                  `mapstructure:"ENV"`
}

var configuration Configuration

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	fmt.Println(envFile)
	viper.SetConfigFile("./.env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile(fmt.Sprintf("../%s", envFile))
		if err := viper.ReadInConfig(); err != nil {
			log.Logger.Printf("Error reading config file \"%s\", %v", envFile, err)
		}
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Logger.Fatal().Msgf("Unable to decode config into map, %v", err)
	}

	fmt.Println("DB url", configuration.Database.DbHost)
}

func GetConfiguration() *Configuration {
	return &configuration
}

func GetRedisConnectionURL() string {
	return configuration.Redis.RedisAddress
}

func GetPoolAddress(poolName string) (string, error) {
	switch poolName {
	case constants.LPCommunity:
		return configuration.Blockchain.LPCommunityPool.LPCommunityAddress, nil
	case constants.LPStaking:
		return configuration.Blockchain.LPStakingPool.LPStakingAddress, nil
	case constants.LPRevenue:
		return configuration.Blockchain.LPRevenuePool.LPRevenueAddress, nil
	case constants.LPTreasury:
		return configuration.Blockchain.LPTreasuryPool.LPTreasuryAddress, nil
	case constants.USDTTreasury:
		return configuration.Blockchain.USDTTreasuryPool.USDTTreasuryAddress, nil
	default:
		return "", fmt.Errorf("unrecognized pool name: %s", poolName)
	}
}

func GetPoolName(poolAddress string) (string, error) {
	switch poolAddress {
	case configuration.Blockchain.LPCommunityPool.LPCommunityAddress:
		return constants.LPCommunity, nil
	case configuration.Blockchain.LPStakingPool.LPStakingAddress:
		return constants.LPStaking, nil
	case configuration.Blockchain.LPRevenuePool.LPRevenueAddress:
		return constants.LPRevenue, nil
	case configuration.Blockchain.LPTreasuryPool.LPTreasuryAddress:
		return constants.LPTreasury, nil
	case configuration.Blockchain.USDTTreasuryPool.USDTTreasuryAddress:
		return constants.USDTTreasury, nil
	default:
		return "", fmt.Errorf("unrecognized pool address: %s", poolAddress)
	}
}
