package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	workertypes "github.com/genefriendway/onchain-handler/internal/workers/types"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type latestBlockWorker struct {
	cacheRepo       cachetypes.CacheRepository
	blockStateUCase ucasetypes.BlockStateUCase
	ethClient       clienttypes.Client
	network         constants.NetworkType
	isRunning       bool       // Tracks if catchup is running
	mu              sync.Mutex // Mutex to protect the isRunning flag
}

func NewLatestBlockWorker(
	cacheRepo cachetypes.CacheRepository,
	blockStateUCase ucasetypes.BlockStateUCase,
	ethClient clienttypes.Client,
	network constants.NetworkType,
) workertypes.Worker {
	return &latestBlockWorker{
		cacheRepo:       cacheRepo,
		blockStateUCase: blockStateUCase,
		ethClient:       ethClient,
		network:         network,
	}
}

// Start starts the periodic task of fetching the latest block and storing it in cache and DB
func (w *latestBlockWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.LatestBlockFetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx)
		case <-ctx.Done():
			logger.GetLogger().Infof("Shutting down latestBlockWorker on network %s", w.network.String())
			return
		}
	}
}

func (w *latestBlockWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warnf("Previous latestBlockWorker on network %s run still in progress, skipping this cycle", w.network.String())
		w.mu.Unlock()
		return
	}

	// Mark as running
	w.isRunning = true
	w.mu.Unlock()

	// Perform the catch-up process
	w.fetchAndStoreLatestBlock(ctx)

	// Mark as not running
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()
}

// fetchAndStoreLatestBlock fetches the latest block and stores it in cache and DB
func (w *latestBlockWorker) fetchAndStoreLatestBlock(ctx context.Context) {
	cacheKey := &cachetypes.Keyer{Raw: constants.LatestBlockCacheKey + w.network.String()}

	// Try to retrieve the latest block from cache
	var latestBlock uint64
	err := w.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err != nil {
		// If cache is empty, load from the database
		latestBlock, err = w.blockStateUCase.GetLatestBlock(ctx, w.network)
		if err != nil {
			logger.GetLogger().Infof("Failed to retrieve latest block from DB: %v", err)
			return
		}

		// Save the latest block to cache for future requests
		err = w.cacheRepo.SaveItem(cacheKey, latestBlock, constants.LatestBlockCacheTime)
		if err != nil {
			logger.GetLogger().Infof("Failed to save latest block to cache: %v", err)
		}
	}

	// Fetch the latest block from the Ethereum blockchain
	blockNumber, err := w.ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		logger.GetLogger().Infof("Failed to fetch latest block from %s: %v", w.network.String(), err)
		return
	}

	// Compare and update if the blockchain block number is newer
	if blockNumber.Uint64() > latestBlock {
		latestBlock = blockNumber.Uint64()

		// Save to cache
		err = w.cacheRepo.SaveItem(cacheKey, latestBlock, constants.LatestBlockCacheTime)
		if err != nil {
			logger.GetLogger().Infof("Failed to save updated block to cache: %v", err)
		}

		// Save to DB
		err = w.blockStateUCase.UpdateLatestBlock(ctx, latestBlock, w.network)
		if err != nil {
			logger.GetLogger().Infof("Failed to update latest block in DB: %v", err)
		}

		logger.GetLogger().Infof("Latest block on network %s updated to: %d", w.network.String(), latestBlock)
	}
}
