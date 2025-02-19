package blockchain

import (
	"context"
	"fmt"

	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// GetLatestBlockFromCacheOrBlockchain retrieves the latest block either from the cache or directly from the blockchain.
func GetLatestBlockFromCacheOrBlockchain(
	ctx context.Context,
	network string,
	cacheRepo cachetypes.CacheRepository,
	ethClient clienttypes.Client,
) (uint64, error) {
	cacheKey := &cachetypes.Keyer{Raw: constants.LatestBlockCacheKey + network}

	var cachedLatestBlock uint64
	err := cacheRepo.RetrieveItem(cacheKey, &cachedLatestBlock)
	if err == nil {
		logger.GetLogger().Debugf("Retrieved %s latest block number from cache: %d", network, cachedLatestBlock)
		return cachedLatestBlock, nil
	}

	// If not found in cache, get the latest block from the blockchain
	latestBlock, err := ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block from blockchain: %w", err)
	}

	logger.GetLogger().Debugf("Retrieved latest block number from %s: %d", network, latestBlock.Uint64())
	return latestBlock.Uint64(), nil
}

// GetLatestBlockFromCache retrieves the latest block from the cache.
func GetLatestBlockFromCache(
	ctx context.Context,
	network string,
	cacheRepo cachetypes.CacheRepository,
) (uint64, error) {
	cacheKey := &cachetypes.Keyer{Raw: constants.LatestBlockCacheKey + network}

	var cachedLatestBlock uint64
	err := cacheRepo.RetrieveItem(cacheKey, &cachedLatestBlock)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block from cache: %w", err)
	}

	logger.GetLogger().Debugf("Retrieved %s latest block number from cache: %d", network, cachedLatestBlock)
	return cachedLatestBlock, nil
}
