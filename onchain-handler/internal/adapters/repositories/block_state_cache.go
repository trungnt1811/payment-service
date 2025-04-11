package repositories

import (
	"context"
	"time"

	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	keyLatestBlock             = "latest_block_"
	keyLastProcessedBlock      = "last_processed_block_"
	keyPrefixBlockState        = "block_state_"
	defaultCacheTimeBlockState = 1 * time.Minute
)

type blockStateCache struct {
	blockStateRepository repotypes.BlockStateRepository
	cache                cachetypes.CacheRepository
}

func NewBlockStateCacheRepository(
	repo repotypes.BlockStateRepository,
	cache cachetypes.CacheRepository,
) repotypes.BlockStateRepository {
	return &blockStateCache{
		blockStateRepository: repo,
		cache:                cache,
	}
}

func (c *blockStateCache) GetLatestBlock(ctx context.Context, network string) (uint64, error) {
	cacheKey := &cachetypes.Keyer{Raw: keyPrefixBlockState + keyLatestBlock + network}

	var cachedValue uint64
	if err := c.cache.RetrieveItem(cacheKey, &cachedValue); err == nil {
		return cachedValue, nil
	}

	block, err := c.blockStateRepository.GetLatestBlock(ctx, network)
	if err != nil {
		return 0, err
	}

	if err := c.cache.SaveItem(cacheKey, block, defaultCacheTimeBlockState); err != nil {
		logger.GetLogger().Warnf("Failed to cache latest block for %s: %v", network, err)
	}

	return block, nil
}

func (c *blockStateCache) UpdateLatestBlock(ctx context.Context, blockNumber uint64, network string) error {
	if err := c.blockStateRepository.UpdateLatestBlock(ctx, blockNumber, network); err != nil {
		return err
	}

	cacheKey := &cachetypes.Keyer{Raw: keyPrefixBlockState + keyLatestBlock + network}
	if err := c.cache.SaveItem(cacheKey, blockNumber, defaultCacheTimeBlockState); err != nil {
		logger.GetLogger().Warnf("Failed to update cache for latest block %s: %v", network, err)
	}

	return nil
}

func (c *blockStateCache) GetLastProcessedBlock(ctx context.Context, network string) (uint64, error) {
	cacheKey := &cachetypes.Keyer{Raw: keyPrefixBlockState + keyLastProcessedBlock + network}

	var cachedValue uint64
	if err := c.cache.RetrieveItem(cacheKey, &cachedValue); err == nil {
		return cachedValue, nil
	}

	block, err := c.blockStateRepository.GetLastProcessedBlock(ctx, network)
	if err != nil {
		return 0, err
	}

	if err := c.cache.SaveItem(cacheKey, block, defaultCacheTimeBlockState); err != nil {
		logger.GetLogger().Warnf("Failed to cache last processed block for %s: %v", network, err)
	}

	return block, nil
}

func (c *blockStateCache) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network string) error {
	if err := c.blockStateRepository.UpdateLastProcessedBlock(ctx, blockNumber, network); err != nil {
		return err
	}

	cacheKey := &cachetypes.Keyer{Raw: keyPrefixBlockState + keyLastProcessedBlock + network}
	if err := c.cache.SaveItem(cacheKey, blockNumber, defaultCacheTimeBlockState); err != nil {
		logger.GetLogger().Warnf("Failed to update cache for last processed block %s: %v", network, err)
	}

	return nil
}
