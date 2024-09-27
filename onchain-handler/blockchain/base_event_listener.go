package blockchain

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

const (
	DefaultEventChannelBufferSize = 100             // Buffer size for event channel
	DefaultBlockOffset            = 10              // Default block offset if last processed block is missing
	MaxBlockRange                 = 2048            // Maximum number of blocks to query at once
	MaxRetries                    = 3               // Maximum number of retries when polling fails
	RetryDelay                    = 3 * time.Second // Delay between retries
)

// BaseEventListener represents the shared behavior of any blockchain event listener.
type BaseEventListener struct {
	ETHClient     *ethclient.Client
	EventChan     chan interface{}
	LastBlockRepo interfaces.BlockStateRepository
	CurrentBlock  uint64
	EventHandlers map[common.Address]func(log types.Log) (interface{}, error) // Map to handle events per contract
}

// NewBaseEventListener initializes a base listener.
func NewBaseEventListener(
	client *ethclient.Client,
	lastBlockRepo interfaces.BlockStateRepository,
	startBlockListener *uint64,
) *BaseEventListener {
	eventChan := make(chan interface{}, DefaultEventChannelBufferSize)

	// Fetch the last processed block from the repository
	lastBlock, err := lastBlockRepo.GetLastProcessedBlock(context.Background())
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

	return &BaseEventListener{
		ETHClient:     client,
		EventChan:     eventChan,
		LastBlockRepo: lastBlockRepo,
		CurrentBlock:  currentBlock, // Store the final determined current block
		EventHandlers: make(map[common.Address]func(log types.Log) (interface{}, error)),
	}
}

// registerEventListener registers an event listener for a specific contract
func (listener *BaseEventListener) registerEventListener(contractAddress string, handler func(log types.Log) (interface{}, error)) {
	address := common.HexToAddress(contractAddress)
	listener.EventHandlers[address] = handler
}

// RunListener starts the listener and processes incoming events.
func (listener *BaseEventListener) RunListener(ctx context.Context) error {
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
	close(listener.EventChan)
	return nil
}

// listen polls the blockchain for logs and parses them.
func (listener *BaseEventListener) listen(ctx context.Context) {
	log.LG.Info("Starting event listener...")

	// Get the last processed block from the repository, defaulting to an offset if not found.
	lastBlock, err := listener.LastBlockRepo.GetLastProcessedBlock(ctx)
	if err != nil || lastBlock == 0 {
		log.LG.Warnf("Failed to get last processed block or it was zero: %v", err)
		latestBlock, err := getLatestBlockNumber(ctx, listener.ETHClient)
		if err != nil {
			log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
			return
		}
		log.LG.Debugf("Retrieved latest block number from blockchain: %d", latestBlock.Uint64())

		if latestBlock.Uint64() > DefaultBlockOffset {
			lastBlock = latestBlock.Uint64() - DefaultBlockOffset
		} else {
			lastBlock = 0
		}
	}

	// Initialize currentBlock based on the stored value
	currentBlock := listener.CurrentBlock
	if currentBlock == 0 {
		currentBlock = lastBlock + 1
	}

	// Continuously listen for new events.
	for {
		// Retrieve the latest block number from the blockchain to stay up-to-date.
		latestBlock, err := getLatestBlockNumber(ctx, listener.ETHClient)
		if err != nil {
			log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
			time.Sleep(RetryDelay)
			continue
		}

		// Ensure we do not go beyond the latest block.
		if currentBlock > latestBlock.Uint64() {
			log.LG.Debugf("No new blocks to process. Waiting for new blocks...")
			time.Sleep(RetryDelay) // Wait before rechecking to prevent excessive polling
			continue
		}

		log.LG.Debugf("Listening for events starting at block: %d", currentBlock)

		// Determine the end block while respecting MaxBlockRange and the latest block.
		endBlock := currentBlock + MaxBlockRange/8
		if endBlock > latestBlock.Uint64() {
			endBlock = latestBlock.Uint64()
		}

		// Extract contract addresses from EventHandlers map
		contractAddresses := make([]common.Address, 0, len(listener.EventHandlers))
		for address := range listener.EventHandlers {
			contractAddresses = append(contractAddresses, address)
		}

		// Process the blocks in chunks of 10 blocks (or DefaultBlockOffset).
		for chunkStart := currentBlock; chunkStart <= endBlock; chunkStart += DefaultBlockOffset {
			chunkEnd := chunkStart + DefaultBlockOffset - 1
			if chunkEnd > endBlock {
				chunkEnd = endBlock
			}

			log.LG.Debugf("Processing block chunk: %d to %d", chunkStart, chunkEnd)

			var logs []types.Log
			// Poll logs from the blockchain with retries in case of failure.
			for retries := 0; retries < MaxRetries; retries++ {
				// Poll logs from the chunk of blocks.
				logs, err = pollForLogsFromBlock(ctx, listener.ETHClient, contractAddresses, chunkStart, chunkEnd)
				if err != nil {
					log.LG.Warnf("Failed to poll logs from block %d to %d: %v. Retrying...", chunkStart, chunkEnd, err)
					time.Sleep(RetryDelay)
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
				if eventHandler, exists := listener.EventHandlers[logEntry.Address]; exists {
					processedEvent, err := eventHandler(logEntry)
					if err != nil {
						log.LG.Warnf("Failed to process log entry: %v", err)
						continue
					}

					// Send the processed event to the channel
					listener.EventChan <- processedEvent
				} else {
					log.LG.Warnf("No event handler for log address: %s", logEntry.Address.Hex())
				}
			}

			// Update the current block for the next iteration.
			currentBlock = chunkEnd + 1
		}

		// Update the last processed block in the repository.
		if err := listener.LastBlockRepo.UpdateLastProcessedBlock(ctx, currentBlock); err != nil {
			log.LG.Errorf("Failed to update last processed block in repository: %v", err)
		}
	}
}

// processEvents handles events from the EventChan.
func (listener *BaseEventListener) processEvents(ctx context.Context) {
	for {
		select {
		case event := <-listener.EventChan:
			log.LG.Debugf("Processed event: %+v", event)

		case <-ctx.Done():
			log.LG.Info("Stopping event processing...")
			return
		}
	}
}
