package conf

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
)

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

func GetRedisConfiguration() *RedisConfiguration {
	return &configuration.Redis
}

func GetWalletConfiguration() *WalletConfiguration {
	return &configuration.Wallet
}

func GetCacheType() string {
	return configuration.CacheType
}

func GetExpiredOrderTime() time.Duration {
	return time.Duration(configuration.PaymentGateway.ExpiredOrderTime) * time.Minute
}

func GetOrderCutoffTime() time.Duration {
	return time.Duration(configuration.PaymentGateway.OrderCutoffTime) * time.Minute
}

func GetPaymentCovering() float64 {
	paymentCoveringStr := configuration.PaymentGateway.PaymentCovering
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

func GetGasBufferMultiplier() float64 {
	multiplierStr := configuration.Blockchain.GasBufferMultiplier
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

func GetTokenSymbol(tokenAddress string) (string, error) {
	tokenSymbols := map[string]string{
		configuration.Blockchain.AvaxNetwork.AvaxUSDTContractAddress: constants.USDT,
		configuration.Blockchain.BscNetwork.BscUSDTContractAddress:   constants.USDT,
	}

	if symbol, exists := tokenSymbols[tokenAddress]; exists {
		return symbol, nil
	}
	return "", fmt.Errorf("unknown token address: %s", tokenAddress)
}

func GetTokenAddress(symbol, network string) (string, error) {
	tokenAddresses := map[string]map[string]string{
		constants.AvaxCChain.String(): {
			constants.USDT: configuration.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		},
		constants.Bsc.String(): {
			constants.USDT: configuration.Blockchain.BscNetwork.BscUSDTContractAddress,
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
