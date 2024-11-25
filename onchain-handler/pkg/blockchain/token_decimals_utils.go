package blockchain

import (
	"context"
	"fmt"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// FetchTokenDecimals retrieves the token decimals for the given token contract address either from the cache or directly from the blockchain.
// If the cache is empty, it saves the token decimals to the cache after retrieving them from the blockchain.
func FetchTokenDecimals(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	tokenContractAddress, network string,
	cacheRepo infrainterfaces.CacheRepository,
) (uint8, error) {
	cacheKey := &caching.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

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
	cacheRepo infrainterfaces.CacheRepository,
) (uint8, error) {
	cacheKey := &caching.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

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
