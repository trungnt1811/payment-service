package listeners

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// baseEventListener represents the shared behavior of any blockchain event listener.
type baseEventListener struct {
	ethClient       pkginterfaces.Client
	network         constants.NetworkType
	eventChan       chan interface{}
	blockStateUCase interfaces.BlockStateUCase
	currentBlock    uint64
	cacheRepo       infrainterfaces.CacheRepository
	eventHandlers   map[common.Address]func(log types.Log) (interface{}, error) // Map to handle events per contract
}

// NewBaseEventListener initializes a base listener.
func NewBaseEventListener(
	client pkginterfaces.Client,
	network constants.NetworkType,
	cacheRepo infrainterfaces.CacheRepository,
	blockStateUCase interfaces.BlockStateUCase,
	startBlockListener *uint64,
) interfaces.BaseEventListener {
	eventChan := make(chan interface{}, constants.DefaultEventChannelBufferSize)

	// Fetch the last processed block from the repository
	lastBlock, err := blockStateUCase.GetLastProcessedBlock(context.Background(), network)
	if err != nil || lastBlock == 0 {
		logger.GetLogger().Warnf("Failed to get last processed block or it was zero: %v", err)
	}

	// Determine the starting block
	currentBlock := lastBlock + 1 // Start at the block after the last processed block

	if startBlockListener != nil && *startBlockListener > lastBlock {
		// Override the current block with the startBlockListener if it's higher than the last processed block
		currentBlock = *startBlockListener
		logger.GetLogger().Debugf("Using startBlockListener on network %s : %d instead of last processed block: %d", string(network), *startBlockListener, lastBlock)
	}

	return &baseEventListener{
		ethClient:       client,
		network:         network,
		cacheRepo:       cacheRepo,
		eventChan:       eventChan,
		blockStateUCase: blockStateUCase,
		currentBlock:    currentBlock, // Store the final determined current block
		eventHandlers:   make(map[common.Address]func(log types.Log) (interface{}, error)),
	}
}

// registerEventListener registers an event listener for a specific contract
func (listener *baseEventListener) RegisterEventListener(contractAddress string, handler func(log types.Log) (interface{}, error)) {
	address := common.HexToAddress(contractAddress)
	listener.eventHandlers[address] = handler
}

// RunListener starts the listener and processes incoming events.
func (listener *baseEventListener) RunListener(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2) // Two goroutines: listen and processEvents

	go func() {
		defer wg.Done()
		listener.listen(ctx)
	}()

	go func() {
		defer wg.Done()
		listener.processEvents(ctx)
	}()

	<-ctx.Done()
	logger.GetLogger().Infof("Event listener on network %s stopped.", string(listener.network))

	// Wait for the goroutines to finish
	wg.Wait()

	// Ensure the channel is closed when the listener stops
	close(listener.eventChan)
	return nil
}

func (listener *baseEventListener) getLatestBlockFromCacheOrBlockchain(ctx context.Context) (uint64, error) {
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey + string(listener.network)}

	var latestBlock uint64
	err := listener.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err == nil {
		logger.GetLogger().Debugf("Retrieved %s latest block number from cache: %d", string(listener.network), latestBlock)
		return latestBlock, nil
	}

	// If cache is empty, load from blockchain
	latest, err := listener.ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve the latest block number from %s: %v", string(listener.network), err)
		return 0, err
	}

	logger.GetLogger().Debugf("Retrieved latest block number from %s: %d", string(listener.network), latest.Uint64())
	return latest.Uint64(), nil
}

// listen polls the blockchain for logs and parses them.
func (listener *baseEventListener) listen(ctx context.Context) {
	logger.GetLogger().Infof("Starting event listener on network %s...", string(listener.network))

	// Get the last processed block from the repository, defaulting to an offset if not found.
	lastBlock, err := listener.blockStateUCase.GetLastProcessedBlock(ctx, listener.network)
	if err != nil || lastBlock == 0 {
		logger.GetLogger().Warnf("Failed to get last processed block on %s or it was zero: %v", string(listener.network), err)

		// Try to retrieve the latest block from cache or blockchain
		latestBlock, err := listener.getLatestBlockFromCacheOrBlockchain(ctx)
		if err != nil {
			logger.GetLogger().Errorf("Failed to retrieve the latest block number from %s: %v", string(listener.network), err)
			return
		}
		logger.GetLogger().Debugf("Retrieved latest block number from network %s: %d", string(listener.network), latestBlock)

		if latestBlock > constants.DefaultBlockOffset {
			lastBlock = latestBlock - constants.DefaultBlockOffset
		} else {
			lastBlock = 0
		}
	}

	// Initialize currentBlock based on the stored value
	currentBlock := listener.currentBlock
	if currentBlock == 0 {
		currentBlock = lastBlock + 1
	}

	// Continuously listen for new events.
	for {
		// Retrieve the latest block number from cache or blockchain to stay up-to-date.
		latestBlock, err := listener.getLatestBlockFromCacheOrBlockchain(ctx)
		if err != nil {
			logger.GetLogger().Errorf("Failed to retrieve the latest block number from %s: %v", string(listener.network), err)
			time.Sleep(constants.RetryDelay)
			continue
		}

		// Calculate the effective latest block considering the confirmation depth.
		effectiveLatestBlock := latestBlock - constants.ConfirmationDepth
		if currentBlock > effectiveLatestBlock {
			logger.GetLogger().Debugf("No new confirmed blocks on network %s to process. Waiting for new blocks...", string(listener.network))
			time.Sleep(constants.RetryDelay) // Wait before rechecking to prevent excessive polling
			continue
		}

		logger.GetLogger().Debugf("Listening for events starting at block on network %s: %d", string(listener.network), currentBlock)

		// Determine the end block while respecting ApiMaxBlocksPerRequest and the effective latest block.
		endBlock := currentBlock + constants.ApiMaxBlocksPerRequest/8
		if endBlock > effectiveLatestBlock {
			endBlock = effectiveLatestBlock
		}

		// Extract contract addresses from EventHandlers map
		var contractAddresses []common.Address
		for address := range listener.eventHandlers {
			contractAddresses = append(contractAddresses, address)
		}

		// Process the blocks in chunks.
		for chunkStart := currentBlock; chunkStart <= endBlock; chunkStart += constants.DefaultBlockOffset {
			chunkEnd := chunkStart + constants.DefaultBlockOffset - 1
			if chunkEnd > endBlock {
				chunkEnd = endBlock
			}

			logger.GetLogger().Debugf("Base Event Listener: Processing block chunk on network %s: %d to %d", string(listener.network), chunkStart, chunkEnd)

			var logs []types.Log
			// Poll logs from the blockchain with retries in case of failure.
			for retries := 0; retries < constants.MaxRetries; retries++ {
				// Poll logs from the chunk of blocks.
				logs, err = listener.ethClient.PollForLogsFromBlock(ctx, contractAddresses, chunkStart, chunkEnd)
				if err != nil {
					logger.GetLogger().Warnf("Failed to poll logs on network %s from block %d to %d: %v. Retrying...", string(listener.network), chunkStart, chunkEnd, err)
					time.Sleep(constants.RetryDelay)
					continue
				}
				break
			}
			if err != nil {
				logger.GetLogger().Errorf("Max retries reached on network %s. Skipping block chunk %d to %d due to error: %v", string(listener.network), chunkStart, chunkEnd, err)
				break // Exit the loop if we cannot fetch logs
			}

			// Apply each parseAndProcessFunc to the logs
			for _, logEntry := range logs {
				if eventHandler, exists := listener.eventHandlers[logEntry.Address]; exists {
					processedEvent, err := eventHandler(logEntry)
					if err != nil {
						logger.GetLogger().Warnf("Failed to process log entry on network %s: %v", string(listener.network), err)
						continue
					}

					// Send the processed event to the channel
					listener.eventChan <- processedEvent
				} else {
					logger.GetLogger().Warnf("No event handler for log address on network %s: %s", string(listener.network), logEntry.Address.Hex())
				}
			}

			// Update the current block for the next iteration.
			currentBlock = chunkEnd + 1
		}

		// Save the last processed block to cache
		cacheKey := &caching.Keyer{Raw: constants.LastProcessedBlockCacheKey + string(listener.network)}
		if err := listener.cacheRepo.SaveItem(cacheKey, latestBlock, constants.LastProcessedBlockCacheTime); err != nil {
			logger.GetLogger().Errorf("Failed to update last processed block on network %s to cache: %v", string(listener.network), err)
		}

		// Update the last processed block in the repository.
		if err := listener.blockStateUCase.UpdateLastProcessedBlock(ctx, currentBlock, listener.network); err != nil {
			logger.GetLogger().Errorf("Failed to update last processed block on network %s in repository: %v", string(listener.network), err)
		}
	}
}

// processEvents handles events from the EventChan.
func (listener *baseEventListener) processEvents(ctx context.Context) {
	for {
		select {
		case event := <-listener.eventChan:
			logger.GetLogger().Debugf("Processed event on network %s: %+v", string(listener.network), event)

		case <-ctx.Done():
			logger.GetLogger().Infof("Stopping event processing on network %s...", string(listener.network))
			return
		}
	}
}
