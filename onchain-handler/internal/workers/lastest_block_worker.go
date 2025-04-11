package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	workertypes "github.com/genefriendway/onchain-handler/internal/workers/types"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type latestBlockWorker struct {
	blockStateUCase ucasetypes.BlockStateUCase
	ethClient       clienttypes.Client
	network         constants.NetworkType
	isRunning       bool       // Tracks if catchup is running
	mu              sync.Mutex // Mutex to protect the isRunning flag
}

func NewLatestBlockWorker(
	blockStateUCase ucasetypes.BlockStateUCase,
	ethClient clienttypes.Client,
	network constants.NetworkType,
) workertypes.Worker {
	return &latestBlockWorker{
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
	// Get the current latest block in DB/cache
	existingBlock, err := w.blockStateUCase.GetLatestBlock(ctx, w.network)
	if err != nil {
		logger.GetLogger().Infof("Failed to get latest block for network %s: %v", w.network.String(), err)
		return
	}

	// Get the current head of the chain
	blockNumber, err := w.ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		logger.GetLogger().Infof("Failed to fetch latest block from %s: %v", w.network.String(), err)
		return
	}

	if blockNumber.Uint64() > existingBlock {
		updatedBlock := blockNumber.Uint64()

		if err := w.blockStateUCase.UpdateLatestBlock(ctx, updatedBlock, w.network); err != nil {
			logger.GetLogger().Infof("Failed to update latest block in DB: %v", err)
			return
		}

		logger.GetLogger().Infof("Latest block on network %s updated to: %d", w.network.String(), updatedBlock)
	}
}
