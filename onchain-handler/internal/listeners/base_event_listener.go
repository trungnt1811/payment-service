package listeners

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/log"
)

// baseEventListener represents the shared behavior of any blockchain event listener.
type baseEventListener struct {
	ethClient       *ethclient.Client
	eventChan       chan interface{}
	blockStateUCase interfaces.BlockStateUCase
	currentBlock    uint64
	cacheRepo       caching.CacheRepository
	eventHandlers   map[common.Address]func(log types.Log) (interface{}, error) // Map to handle events per contract
}

// NewBaseEventListener initializes a base listener.
func NewBaseEventListener(
	client *ethclient.Client,
	cacheRepo caching.CacheRepository,
	blockStateUCase interfaces.BlockStateUCase,
	startBlockListener *uint64,
) interfaces.BaseEventListener {
	eventChan := make(chan interface{}, constants.DefaultEventChannelBufferSize)

	// Fetch the last processed block from the repository
	lastBlock, err := blockStateUCase.GetLastProcessedBlock(context.Background())
	if err != nil || lastBlock == 0 {
		log.LG.Warnf("Failed to get last processed block or it was zero: %v", err)
	}

	// Determine the starting block
	currentBlock := lastBlock + 1 // Start at the block after the last processed block

	if startBlockListener != nil && *startBlockListener > lastBlock {
		// Override the current block with the startBlockListener if it's higher than the last processed block
		currentBlock = *startBlockListener
		log.LG.Debugf("Using startBlockListener: %d instead of last processed block: %d", *startBlockListener, lastBlock)
	}

	return &baseEventListener{
		ethClient:       client,
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
	log.LG.Info("Event listener stopped.")

	// Wait for the goroutines to finish
	wg.Wait()

	// Ensure the channel is closed when the listener stops
	close(listener.eventChan)
	return nil
}

func (listener *baseEventListener) getLatestBlockFromCacheOrBlockchain(ctx context.Context) (uint64, error) {
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey}

	var latestBlock uint64
	err := listener.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err == nil {
		log.LG.Debugf("Retrieved latest block number from cache: %d", latestBlock)
		return latestBlock, nil
	}

	// If cache is empty, load from blockchain
	latest, err := utils.GetLatestBlockNumber(ctx, listener.ethClient)
	if err != nil {
		log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
		return 0, err
	}

	log.LG.Debugf("Retrieved latest block number from blockchain: %d", latest.Uint64())
	return latest.Uint64(), nil
}

// listen polls the blockchain for logs and parses them.
func (listener *baseEventListener) listen(ctx context.Context) {
	log.LG.Info("Starting event listener...")

	// Get the last processed block from the repository, defaulting to an offset if not found.
	lastBlock, err := listener.blockStateUCase.GetLastProcessedBlock(ctx)
	if err != nil || lastBlock == 0 {
		log.LG.Warnf("Failed to get last processed block or it was zero: %v", err)

		// Try to retrieve the latest block from cache or blockchain
		latestBlock, err := listener.getLatestBlockFromCacheOrBlockchain(ctx)
		if err != nil {
			log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
			return
		}
		log.LG.Debugf("Retrieved latest block number: %d", latestBlock)

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
			log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
			time.Sleep(constants.RetryDelay)
			continue
		}

		// Calculate the effective latest block considering the confirmation depth.
		effectiveLatestBlock := latestBlock - constants.ConfirmationDepth
		if currentBlock > effectiveLatestBlock {
			log.LG.Debugf("No new confirmed blocks to process. Waiting for new blocks...")
			time.Sleep(constants.RetryDelay) // Wait before rechecking to prevent excessive polling
			continue
		}

		log.LG.Debugf("Listening for events starting at block: %d", currentBlock)

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

			log.LG.Debugf("Base Event Listener: Processing block chunk: %d to %d", chunkStart, chunkEnd)

			var logs []types.Log
			// Poll logs from the blockchain with retries in case of failure.
			for retries := 0; retries < constants.MaxRetries; retries++ {
				// Poll logs from the chunk of blocks.
				logs, err = utils.PollForLogsFromBlock(ctx, listener.ethClient, contractAddresses, chunkStart, chunkEnd)
				if err != nil {
					log.LG.Warnf("Failed to poll logs from block %d to %d: %v. Retrying...", chunkStart, chunkEnd, err)
					time.Sleep(constants.RetryDelay)
					continue
				}
				break
			}
			if err != nil {
				log.LG.Errorf("Max retries reached. Skipping block chunk %d to %d due to error: %v", chunkStart, chunkEnd, err)
				break // Exit the loop if we cannot fetch logs
			}

			// Apply each parseAndProcessFunc to the logs
			for _, logEntry := range logs {
				if eventHandler, exists := listener.eventHandlers[logEntry.Address]; exists {
					processedEvent, err := eventHandler(logEntry)
					if err != nil {
						log.LG.Warnf("Failed to process log entry: %v", err)
						continue
					}

					// Send the processed event to the channel
					listener.eventChan <- processedEvent
				} else {
					log.LG.Warnf("No event handler for log address: %s", logEntry.Address.Hex())
				}
			}

			// Update the current block for the next iteration.
			currentBlock = chunkEnd + 1
		}

		// Update the last processed block in the repository.
		if err := listener.blockStateUCase.UpdateLastProcessedBlock(ctx, currentBlock); err != nil {
			log.LG.Errorf("Failed to update last processed block in repository: %v", err)
		}
	}
}

// processEvents handles events from the EventChan.
func (listener *baseEventListener) processEvents(ctx context.Context) {
	for {
		select {
		case event := <-listener.eventChan:
			log.LG.Debugf("Processed event: %+v", event)

		case <-ctx.Done():
			log.LG.Info("Stopping event processing...")
			return
		}
	}
}
