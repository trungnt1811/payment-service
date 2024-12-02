package listeners

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// tokenTransferListener listens for token transfers and processes them using a queue of payment orders.
type tokenTransferListener struct {
	ctx                      context.Context
	config                   *conf.Configuration
	cacheRepo                infrainterfaces.CacheRepository
	baseEventListener        interfaces.BaseEventListener
	paymentOrderUCase        interfaces.PaymentOrderUCase
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase
	network                  constants.NetworkType
	tokenContractAddress     string
	tokenDecimals            uint8
	parsedABI                abi.ABI
	queue                    *queue.Queue[dto.PaymentOrderDTO]
	mu                       sync.Mutex // Mutex for ticker synchronization
}

// NewTokenTransferListener creates a new tokenTransferListener with a payment order queue.
func NewTokenTransferListener(
	ctx context.Context,
	config *conf.Configuration,
	cacheRepo infrainterfaces.CacheRepository,
	baseEventListener interfaces.BaseEventListener,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	network constants.NetworkType,
	tokenContractAddress string,
	orderQueue *queue.Queue[dto.PaymentOrderDTO],
) (interfaces.EventListener, error) {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	decimals, err := blockchain.GetTokenDecimalsFromCache(tokenContractAddress, string(network), cacheRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to get token decimals: %w", err)
	}

	listener := &tokenTransferListener{
		ctx:                      ctx,
		config:                   config,
		cacheRepo:                cacheRepo,
		baseEventListener:        baseEventListener,
		paymentOrderUCase:        paymentOrderUCase,
		paymentEventHistoryUCase: paymentEventHistoryUCase,
		network:                  network,
		tokenContractAddress:     tokenContractAddress,
		tokenDecimals:            decimals,
		queue:                    orderQueue,
		parsedABI:                parsedABI,
	}

	go listener.startDequeueTicker(constants.DequeueInterval)

	return listener, nil
}

// startDequeueTicker starts a ticker that triggers dequeueing expired or successful orders at a specified interval.
func (listener *tokenTransferListener) startDequeueTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to lock before dequeueing, ensuring no overlap
			listener.mu.Lock()
			logger.GetLogger().Debug("Starting dequeue operation...")
			listener.dequeueOrders()
			logger.GetLogger().Debug("Dequeue operation finished.")
			listener.mu.Unlock()
		case <-listener.ctx.Done():
			logger.GetLogger().Debugf("Stopping dequeue ticker: %v", listener.ctx.Err())
			return
		}
	}
}

func (listener *tokenTransferListener) parseAndProcessRealtimeTransferEvent(vLog types.Log) (interface{}, error) {
	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := listener.config.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack realtime transfer event on network %s for address %s, block number %d: %w", string(listener.network), vLog.Address.Hex(), vLog.BlockNumber, err)
	}

	logger.GetLogger().Infof("Detected realtime transfer event on network %s : From: %s, To: %s, Value: %s", string(listener.network), transferEvent.From.Hex(), transferEvent.To.Hex(), transferEvent.Value.String())

	queueItems := listener.queue.GetItems()

	// Process each payment order based on the transfer event
	for _, order := range queueItems {
		// Check if the transfer matches the order's wallet and token symbol
		if strings.EqualFold(transferEvent.To.Hex(), order.PaymentAddress) && strings.EqualFold(order.Symbol, tokenSymbol) {
			err := listener.paymentOrderUCase.UpdateOrderStatus(listener.ctx, order.ID, constants.Processing)
			if err != nil {
				logger.GetLogger().Errorf("Failed to update order status to processing on network %s for order ID %d, error: %v", string(listener.network), order.ID, err)
				return nil, err
			}
		}
	}

	return transferEvent, nil
}

// parseAndProcessConfirmedTransferEvent parses and processes a confirmed transfer event, checking if it matches any payment order in the queue.
func (listener *tokenTransferListener) parseAndProcessConfirmedTransferEvent(vLog types.Log) (interface{}, error) {
	// Handle expired and successful orders before processing the event
	listener.dequeueOrders()

	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := listener.config.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack confirmed transfer event on network %s for address %s, block number %d: %w", string(listener.network), vLog.Address.Hex(), vLog.BlockNumber, err)
	}

	logger.GetLogger().Infof("Detected confirmed transfer event on network %s : From: %s, To: %s, Value: %s", string(listener.network), transferEvent.From.Hex(), transferEvent.To.Hex(), transferEvent.Value.String())

	queueItems := listener.queue.GetItems()

	var processedOrder dto.PaymentOrderDTOResponse

	// Process each payment order based on the transfer event
	for index, order := range queueItems {
		// Check if the transfer matches the order's wallet and token symbol
		if strings.EqualFold(transferEvent.To.Hex(), order.PaymentAddress) && strings.EqualFold(order.Symbol, tokenSymbol) {
			// Convert transfer event value from Wei to Eth
			transferEventValueInEth, err := utils.ConvertSmallestUnitToFloatToken(transferEvent.Value.String(), listener.tokenDecimals)
			if err != nil {
				logger.GetLogger().Errorf("Failed to convert transfer event on network %s value to ETH for order ID %d, error: %v", string(listener.network), order.ID, err)
				return nil, err
			}

			// Prepare the payload for the payment event history
			payloads := []dto.PaymentEventPayloadDTO{
				{
					PaymentOrderID:  order.ID,
					TransactionHash: vLog.TxHash.Hex(),
					FromAddress:     transferEvent.From.Hex(),
					ToAddress:       transferEvent.To.Hex(),
					ContractAddress: vLog.Address.Hex(),
					TokenSymbol:     tokenSymbol,
					Amount:          transferEventValueInEth,
					Network:         string(listener.network),
				},
			}

			// Create payment event history
			if err := listener.paymentEventHistoryUCase.CreatePaymentEventHistory(listener.ctx, payloads); err != nil {
				logger.GetLogger().Errorf("Failed to process payment event history on network %s for order ID %d, error: %v", string(listener.network), order.ID, err)
				return nil, err
			}

			// Process the payment for the order
			if err := listener.processOrderPayment(index, order, transferEvent, vLog.BlockNumber); err != nil {
				logger.GetLogger().Errorf("Failed to process payment on network %s for order ID %d, error: %v", string(listener.network), order.ID, err)
				return nil, err
			}

			// Get the processed order
			processedOrder, err = listener.paymentOrderUCase.GetPaymentOrderByID(listener.ctx, order.ID)
			if err != nil {
				logger.GetLogger().Errorf("Failed to get the processed order by ID %d, error: %v", order.ID, err)
				return nil, err
			}
			break
		}
	}

	return processedOrder, nil
}

// processOrderPayment handles the payment for an order based on the transfer event details.
// It updates the order status and wallet usage based on the payment amount.
func (listener *tokenTransferListener) processOrderPayment(itemIndex int, order dto.PaymentOrderDTO, transferEvent blockchain.TransferEvent, blockHeight uint64) error {
	orderAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Amount, listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := payment.CalculatePaymentCovering(orderAmount, listener.config.GetPaymentCovering())

	transferredAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Transferred, listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("failed to convert transferred amount: %v", err)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred := new(big.Int).Add(transferredAmount, transferEvent.Value)

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		logger.GetLogger().Infof("Processed full payment on network %s for order ID: %d", string(listener.network), order.ID)

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return listener.updatePaymentOrderStatus(itemIndex, order, constants.Success, totalTransferred.String(), blockHeight)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		logger.GetLogger().Infof("Processed partial payment on network %s for order ID: %d", string(listener.network), order.ID)

		// Check if the order is still 'Processing' or needs to be marked as 'Partial'.
		status := constants.Partial
		if blockHeight < order.BlockHeight {
			status = constants.Processing
		}

		// Update the order status and keep the wallet associated with the order.
		return listener.updatePaymentOrderStatus(itemIndex, order, status, totalTransferred.String(), blockHeight)
	}

	return nil
}

func (listener *tokenTransferListener) updatePaymentOrderStatus(
	itemIndex int,
	order dto.PaymentOrderDTO,
	status, transferredAmount string,
	blockHeight uint64,
) error {
	// Convert transferredAmount from Wei to Eth
	transferredAmountInEth, err := utils.ConvertSmallestUnitToFloatToken(transferredAmount, listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("updatePaymentOrderStatus error: %v", err)
	}

	// Also update item in queue
	order.Transferred = transferredAmountInEth
	order.Status = status
	if err := listener.queue.ReplaceItemAtIndex(itemIndex, order); err != nil {
		return fmt.Errorf("failed to update order in queue: %w", err)
	}

	// Call the method to update the payment order
	return listener.paymentOrderUCase.UpdatePaymentOrder(listener.ctx, order.ID, &blockHeight, &status, &transferredAmountInEth, nil)
}

// dequeueOrders removes expired or successful orders from the queue and refills it to maintain the limit.
func (listener *tokenTransferListener) dequeueOrders() {
	// Retrieve last processed block from cache
	cacheKey := &caching.Keyer{Raw: constants.LastProcessedBlockCacheKey + string(listener.network)}

	var latestProcessedBlock uint64
	err := listener.cacheRepo.RetrieveItem(cacheKey, &latestProcessedBlock)
	if err == nil {
		logger.GetLogger().Debugf("Retrieved %s last processed block from cache: %d", string(listener.network), latestProcessedBlock)
		latestProcessedBlock = 0
	}

	// Retrieve all current orders in the queue
	orders := listener.queue.GetItems()

	var dequeueOrders []dto.PaymentOrderDTO
	// Iterate over current orders in the queue
	for index, order := range orders {
		// Check if the order needs to be dequeued
		if listener.shouldDequeueOrder(order) {
			if order.Status != constants.Success {
				orders[index].Status = constants.Expired
				if latestProcessedBlock > 0 {
					orders[index].BlockHeight = latestProcessedBlock
				}
				dequeueOrders = append(dequeueOrders, orders[index])
			}
			if err := listener.queue.Dequeue(func(o dto.PaymentOrderDTO) bool {
				return o.ID == order.ID
			}); err != nil {
				logger.GetLogger().Errorf("Failed to dequeue order ID: %d, error: %v", order.ID, err)
				return
			}
		}
	}

	// Update the statuses of dequeued orders
	if len(dequeueOrders) > 0 {
		err := listener.paymentOrderUCase.BatchUpdateOrderStatuses(listener.ctx, dequeueOrders)
		// Log the result of the batch update
		if err != nil {
			// Log all failed order IDs in case of an error
			var failedOrderIDs []uint64
			for _, order := range dequeueOrders {
				failedOrderIDs = append(failedOrderIDs, order.ID)
			}
			logger.GetLogger().Errorf("Failed to update orders with IDs: %v, error: %v", failedOrderIDs, err)
		} else {
			// Log all successful order IDs
			var successfulOrderIDs []uint64
			for _, order := range dequeueOrders {
				successfulOrderIDs = append(successfulOrderIDs, order.ID)
			}
			logger.GetLogger().Infof("Successfully updated orders with IDs: %v to expired status", successfulOrderIDs)
		}
	}

	// Refill the queue to ensure it has the required number of items
	if err := listener.queue.FillQueue(); err != nil {
		logger.GetLogger().Errorf("Failed to refill the %s queue: %v", string(listener.network), err)
	}
}

// shouldDequeueOrder checks if an order is expired or succeeded.
func (listener *tokenTransferListener) shouldDequeueOrder(order dto.PaymentOrderDTO) bool {
	return listener.isOrderExpired(order) || listener.isOrderSucceeded(order)
}

// isOrderExpired check if the order has expired
func (listener *tokenTransferListener) isOrderExpired(order dto.PaymentOrderDTO) bool {
	currentTime := time.Now().UTC()
	expiredTime := order.ExpiredTime.UTC()
	return currentTime.After(expiredTime)
}

func (listener *tokenTransferListener) isOrderSucceeded(order dto.PaymentOrderDTO) bool {
	return order.Status == constants.Success
}

// Register registers the listeners for token transfer events.
func (listener *tokenTransferListener) Register(ctx context.Context) {
	listener.baseEventListener.RegisterConfirmedEventListener(
		listener.tokenContractAddress,
		listener.parseAndProcessConfirmedTransferEvent,
	)
	listener.baseEventListener.RegisterRealtimeEventListener(
		listener.tokenContractAddress,
		listener.parseAndProcessRealtimeTransferEvent,
	)
}
