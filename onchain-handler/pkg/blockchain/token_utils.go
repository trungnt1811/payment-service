package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// FetchTokenDecimals retrieves the token decimals for the given token contract address either from the cache or directly from the blockchain.
// If the cache is empty, it saves the token decimals to the cache after retrieving them from the blockchain.
func FetchTokenDecimals(
	ctx context.Context,
	ethClient clienttypes.Client,
	tokenContractAddress, network string,
	cacheRepo cachetypes.CacheRepository,
) (uint8, error) {
	cacheKey := &cachetypes.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

	var cachedDecimals uint8
	err := cacheRepo.RetrieveItem(cacheKey, &cachedDecimals)
	if err == nil {
		logger.GetLogger().Debugf(
			"Token decimals retrieved from cache. Token: %s, Network: %s, Decimals: %d",
			tokenContractAddress,
			network,
			cachedDecimals,
		)
		return cachedDecimals, nil
	}

	// If not found in cache, get the token decimals from the blockchain
	decimals, err := ethClient.GetTokenDecimals(ctx, tokenContractAddress)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to get token decimals from blockchain. Token: %s, Network: %s, Error: %w",
			tokenContractAddress,
			network,
			err,
		)
	}

	logger.GetLogger().Debugf(
		"Token decimals retrieved from blockchain. Token: %s, Network: %s, Decimals: %d",
		tokenContractAddress,
		network,
		decimals,
	)

	// Save the retrieved decimals to the cache
	// Duration -1 means the item will never expire
	err = cacheRepo.SaveItem(cacheKey, decimals, -1)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to persist token decimals to cache. Token: %s, Network: %s, Error: %v",
			tokenContractAddress,
			network,
			err,
		)
	}

	return decimals, nil
}

func GetTokenDecimalsFromCache(
	tokenContractAddress, network string,
	cacheRepo cachetypes.CacheRepository,
) (uint8, error) {
	cacheKey := &cachetypes.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

	var cachedDecimals uint8
	err := cacheRepo.RetrieveItem(cacheKey, &cachedDecimals)
	if err == nil {
		logger.GetLogger().Debugf(
			"Token decimals retrieved from cache. Token: %s, Network: %s, Decimals: %d",
			tokenContractAddress,
			network,
			cachedDecimals,
		)
		return cachedDecimals, nil
	}

	return 0, fmt.Errorf("token decimals not found in cache. Token: %s, Network: %s", tokenContractAddress, network)
}

// GetNativeTokenSymbol returns the symbol of the native token for the given network
func GetNativeTokenSymbol(network constants.NetworkType) (string, error) {
	switch network {
	case constants.AvaxCChain:
		return "AVAX", nil
	case constants.Bsc:
		return "BNB", nil
	default:
		return "", fmt.Errorf("unsupported network type: %s", network)
	}
}

// FetchTokenBalance retrieves the token balance for the given wallet address and token contract address either from the cache or directly from the blockchain.
func FetchTokenBalance(
	ctx context.Context,
	ethClient clienttypes.Client,
	tokenContractAddress, walletAddress, network string,
	cacheRepo cachetypes.CacheRepository,
) (*big.Float, error) {
	// Step 1: Retrieve the raw token balance from the blockchain
	rawBalance, err := ethClient.GetTokenBalance(ctx, tokenContractAddress, walletAddress)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get token balance from blockchain. Wallet: %s, Token: %s, Network: %s, Error: %w",
			walletAddress,
			tokenContractAddress,
			network,
			err,
		)
	}

	// Step 2: Retrieve token decimals from the cache
	decimals, err := GetTokenDecimalsFromCache(tokenContractAddress, network, cacheRepo)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve token decimals from cache. Token: %s, Network: %s, Error: %w",
			tokenContractAddress,
			network,
			err,
		)
	}

	// Step 3: Convert the raw balance to a human-readable format
	factor := new(big.Float).SetFloat64(float64(10 ^ decimals))
	humanReadableBalance := new(big.Float).Quo(new(big.Float).SetInt(rawBalance), factor)

	logger.GetLogger().Infof(
		"Token balance retrieved successfully. Wallet: %s, Token: %s, Network: %s, Balance: %s",
		walletAddress,
		tokenContractAddress,
		network,
		humanReadableBalance.String(),
	)

	return humanReadableBalance, nil
}
