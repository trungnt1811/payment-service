package workers

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/log"
)

type latestBlockWorker struct {
	cacheRepo       infrainterfaces.CacheRepository
	blockStateUCase interfaces.BlockStateUCase
	ethClient       *ethclient.Client
	network         constants.NetworkType
	isRunning       bool       // Tracks if catchup is running
	mu              sync.Mutex // Mutex to protect the isRunning flag
}

func NewLatestBlockWorker(
	cacheRepo infrainterfaces.CacheRepository,
	blockStateUCase interfaces.BlockStateUCase,
	ethClient *ethclient.Client,
	network constants.NetworkType,
) interfaces.Worker {
	return &latestBlockWorker{
		cacheRepo:       cacheRepo,
		blockStateUCase: blockStateUCase,
		ethClient:       ethClient,
		network:         network,
	}
}

// Start starts the periodic task of fetching the latest block and storing it in cache and DB
func (w *latestBlockWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.LatestBlockFetchingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx)
		case <-ctx.Done():
			log.GetLogger().Infof("Shutting down latestBlockWorker on network %s", string(w.network))
			return
		}
	}
}

func (w *latestBlockWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		log.GetLogger().Warnf("Previous latestBlockWorker on network %s run still in progress, skipping this cycle", string(w.network))
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
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey + string(w.network)}

	// Try to retrieve the latest block from cache
	var latestBlock uint64
	err := w.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err != nil {
		// If cache is empty, load from the database
		latestBlock, err = w.blockStateUCase.GetLatestBlock(ctx, w.network)
		if err != nil {
			log.GetLogger().Infof("Failed to retrieve latest block from DB: %v", err)
			return
		}

		// Save the latest block to cache for future requests
		err = w.cacheRepo.SaveItem(cacheKey, latestBlock, constants.LatestBlockCacheTime)
		if err != nil {
			log.GetLogger().Infof("Failed to save latest block to cache: %v", err)
		}
	}

	// Fetch the latest block from the Ethereum blockchain
	blockNumber, err := utils.GetLatestBlockNumber(ctx, w.ethClient)
	if err != nil {
		log.GetLogger().Infof("Failed to fetch latest block from %s: %v", string(w.network), err)
		return
	}

	// Compare and update if the blockchain block number is newer
	if blockNumber.Uint64() > latestBlock {
		latestBlock = blockNumber.Uint64()

		// Save to cache
		err = w.cacheRepo.SaveItem(cacheKey, latestBlock, constants.LatestBlockCacheTime)
		if err != nil {
			log.GetLogger().Infof("Failed to save updated block to cache: %v", err)
		}

		// Save to DB
		err = w.blockStateUCase.UpdateLatestBlock(ctx, latestBlock, w.network)
		if err != nil {
			log.GetLogger().Infof("Failed to update latest block in DB: %v", err)
		}
	}

	log.GetLogger().Infof("Latest block on network %s updated to: %d", string(w.network), latestBlock)
}
