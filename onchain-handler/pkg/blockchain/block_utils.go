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

// GetLatestBlockFromCacheOrBlockchain retrieves the latest block either from the cache or directly from the blockchain.
func GetLatestBlockFromCacheOrBlockchain(
	ctx context.Context,
	network string,
	cacheRepo infrainterfaces.CacheRepository,
	ethClient pkginterfaces.Client,
) (uint64, error) {
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey + network}

	var latestBlock uint64
	err := cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err == nil {
		logger.GetLogger().Debugf("Retrieved %s latest block number from cache: %d", network, latestBlock)
		return latestBlock, nil
	}

	// If not found in cache, get the latest block from the blockchain
	latest, err := ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block from blockchain: %w", err)
	}

	logger.GetLogger().Debugf("Retrieved latest block number from %s: %d", network, latest.Uint64())
	return latest.Uint64(), nil
}
