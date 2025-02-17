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
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// tokenTransferListener listens for token transfers and processes them using a queue of payment orders.
type tokenTransferListener struct {
	ctx                      context.Context
	cacheRepo                infrainterfaces.CacheRepository
	baseEventListener        interfaces.BaseEventListener
	paymentOrderUCase        interfaces.PaymentOrderUCase
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase
	paymentStatisticsUCase   interfaces.PaymentStatisticsUCase
	paymentWalletUCase       interfaces.PaymentWalletUCase
	network                  constants.NetworkType
	tokenContractAddress     string
	tokenDecimals            uint8
	parsedABI                abi.ABI
	queue                    infrainterfaces.Queue[dto.PaymentOrderDTO]
	mu                       sync.Mutex // Mutex for ticker synchronization
}

// NewTokenTransferListener creates a new tokenTransferListener with a payment order queue.
func NewTokenTransferListener(
	ctx context.Context,
	cacheRepo infrainterfaces.CacheRepository,
	baseEventListener interfaces.BaseEventListener,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	network constants.NetworkType,
	tokenContractAddress string,
	orderQueue infrainterfaces.Queue[dto.PaymentOrderDTO],
) (interfaces.EventListener, error) {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	decimals, err := blockchain.GetTokenDecimalsFromCache(tokenContractAddress, network.String(), cacheRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to get token decimals: %w", err)
	}

	listener := &tokenTransferListener{
		ctx:                      ctx,
		cacheRepo:                cacheRepo,
		baseEventListener:        baseEventListener,
		paymentOrderUCase:        paymentOrderUCase,
		paymentEventHistoryUCase: paymentEventHistoryUCase,
		paymentStatisticsUCase:   paymentStatisticsUCase,
		paymentWalletUCase:       paymentWalletUCase,
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
			// Refill the queue to ensure it has the required number of items
			if err := listener.queue.FillQueue(); err != nil {
				logger.GetLogger().Errorf("Failed to refill the %s queue: %v", listener.network.String(), err)
			}
			logger.GetLogger().Debug("Dequeue operation finished.")
			listener.mu.Unlock()
		case <-listener.ctx.Done():
			logger.GetLogger().Debugf("Stopping dequeue ticker: %v", listener.ctx.Err())
			return
		}
	}
}

func (listener *tokenTransferListener) fetchOrderDetailsFromQueue(transferEvent blockchain.TransferEvent, tokenSymbol string) (*dto.PaymentOrderDTO, error) {
	key := transferEvent.To.Hex() + "_" + tokenSymbol
	index, exists := listener.queue.GetIndex(key)
	if !exists {
		return nil, nil // No matching order
	}

	order, err := listener.queue.GetItemAtIndex(index)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve order from queue at index %d: %v", index, err)
		return nil, err
	}

	orderDTO, err := listener.paymentOrderUCase.GetPaymentOrderByID(listener.ctx, order.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch order ID %d from DB: %v", order.ID, err)
		return nil, err
	}

	// Ensure order matches the correct network
	if orderDTO.Network != listener.network.String() {
		return nil, nil
	}

	return &order, nil
}

func (listener *tokenTransferListener) parseAndProcessRealtimeTransferEvent(vLog types.Log) (interface{}, error) {
	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := conf.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack realtime transfer event on network %s for address %s, block number %d: %w", listener.network.String(), vLog.Address.Hex(), vLog.BlockNumber, err)
	}

	// Fetch the order details from the queue
	order, err := listener.fetchOrderDetailsFromQueue(transferEvent, tokenSymbol)
	if err != nil || order == nil {
		return nil, err
	}

	// Update the order status to 'Processing'
	status := constants.Processing
	upcomingBlockHeight := vLog.BlockNumber
	err = listener.paymentOrderUCase.UpdatePaymentOrder(listener.ctx, order.ID, nil, &upcomingBlockHeight, &status, nil, nil)
	if err != nil {
		logger.GetLogger().Errorf("Failed to update order status to processing on network %s for order ID %d, error: %v", listener.network.String(), order.ID, err)
		return nil, err
	}
	logger.GetLogger().Infof("Updated order ID %d to status 'Processing' on network %s", order.ID, listener.network.String())

	// Recheck index before updating the queue
	key := transferEvent.To.Hex() + "_" + tokenSymbol
	newIndex, stillExists := listener.queue.GetIndex(key)
	if !stillExists {
		logger.GetLogger().Warnf("Order ID %d was removed from the queue before it could be updated", order.ID)
		return nil, fmt.Errorf("order no longer exists in the queue")
	}

	// Update the order status in the queue
	order.Status = status
	order.UpcomingBlockHeight = upcomingBlockHeight
	if err := listener.queue.ReplaceItemAtIndex(newIndex, *order); err != nil {
		logger.GetLogger().Errorf("Failed to update order in queue: %v", err)
		return nil, err
	}

	return transferEvent, nil
}

// parseAndProcessConfirmedTransferEvent parses and processes a confirmed transfer event, checking if it matches any payment order in the queue.
func (listener *tokenTransferListener) parseAndProcessConfirmedTransferEvent(vLog types.Log) (interface{}, error) {
	// Handle expired and successful orders before processing the event
	listener.dequeueOrders()

	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := conf.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack confirmed transfer event on network %s for address %s, block number %d: %w", listener.network.String(), vLog.Address.Hex(), vLog.BlockNumber, err)
	}

	// Fetch the order details from the queue
	order, err := listener.fetchOrderDetailsFromQueue(transferEvent, tokenSymbol)
	if err != nil || order == nil {
		return nil, err
	}

	// Convert transfer amount to token units
	transferEventValueInEth, err := utils.ConvertSmallestUnitToFloatToken(transferEvent.Value.String(), listener.tokenDecimals)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to convert transfer event on network %s value to ETH for order ID %d, error: %v",
			listener.network.String(), order.ID, err,
		)
		return nil, err
	}

	// Prepare payment event history payload
	payload := dto.PaymentEventPayloadDTO{
		PaymentOrderID:  order.ID,
		TransactionHash: vLog.TxHash.Hex(),
		FromAddress:     transferEvent.From.Hex(),
		ToAddress:       transferEvent.To.Hex(),
		ContractAddress: vLog.Address.Hex(),
		TokenSymbol:     tokenSymbol,
		Amount:          transferEventValueInEth,
		Network:         listener.network.String(),
	}

	// Process Order Payment
	isUpdated, err := listener.processOrderPayment(*order, transferEvent, vLog.BlockNumber)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to process payment on network %s for order ID %d, error: %v",
			listener.network.String(), order.ID, err,
		)
		return nil, err
	}
	if isUpdated {
		// Store payment event history
		if err := listener.paymentEventHistoryUCase.CreatePaymentEventHistory(
			listener.ctx, []dto.PaymentEventPayloadDTO{payload},
		); err != nil {
			logger.GetLogger().Errorf("Failed to store payment event history on network %s for order ID %d, error: %v",
				listener.network.String(), order.ID, err)
			return nil, err
		}

		// Update Payment Statistics & Wallet Balance
		granularity := constants.Daily
		periodStart := utils.GetPeriodStart(granularity, time.Now())

		if err := listener.paymentStatisticsUCase.IncrementStatistics(
			listener.ctx, granularity, periodStart, nil, &payload.Amount, payload.TokenSymbol, order.VendorID,
		); err != nil {
			logger.GetLogger().Errorf("Failed to increment payment statistics on network %s for order ID %d, error: %v",
				listener.network.String(), order.ID, err)
		}

		if err := listener.paymentWalletUCase.AddPaymentWalletBalance(
			listener.ctx, order.Wallet.ID, payload.Amount, listener.network, payload.TokenSymbol,
		); err != nil {
			logger.GetLogger().Errorf("Failed to add payment wallet balance on network %s for order ID %d, error: %v",
				listener.network.String(), order.ID, err)
		}
	}

	// Retrieve updated order from DB
	processedOrder, err := listener.paymentOrderUCase.GetPaymentOrderByID(listener.ctx, order.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get the processed order by ID %d, error: %v", order.ID, err)
		return nil, err
	}

	// Return nil if no order was processed
	if processedOrder.ID == 0 {
		return nil, nil
	}

	return processedOrder, nil
}

// processOrderPayment handles the payment for an order based on the transfer event details.
// It updates the order status and wallet usage based on the payment amount.
func (listener *tokenTransferListener) processOrderPayment(
	order dto.PaymentOrderDTO,
	transferEvent blockchain.TransferEvent,
	blockHeight uint64,
) (bool, error) {
	orderAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Amount, listener.tokenDecimals)
	if err != nil {
		return false, fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := payment.CalculatePaymentCoveringAsDiscount(
		orderAmount, conf.GetPaymentCovering(), listener.tokenDecimals,
	)

	// Get newest order state in cache
	orderDTO, err := listener.paymentOrderUCase.GetPaymentOrderByID(listener.ctx, order.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get order by ID %d, error: %v", order.ID, err)
		return false, err
	}

	transferredAmount, err := utils.ConvertFloatTokenToSmallestUnit(orderDTO.Transferred, listener.tokenDecimals)
	if err != nil {
		return false, fmt.Errorf("failed to convert transferred amount: %v", err)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred := new(big.Int).Add(transferredAmount, transferEvent.Value)

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		logger.GetLogger().Infof("Processed full payment on network %s for order ID: %d", listener.network.String(), order.ID)

		// Check if order is still 'Processing' or needs to be marked as 'Success'.
		status := constants.Success
		if blockHeight < orderDTO.UpcomingBlockHeight {
			status = constants.Processing
		}

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return true, listener.updatePaymentOrderStatus(order, status, totalTransferred.String(), blockHeight)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		logger.GetLogger().Infof("Processed partial payment on network %s for order ID: %d", listener.network.String(), order.ID)

		// Check if the order is still 'Processing' or needs to be marked as 'Partial'.
		status := constants.Partial
		if blockHeight < orderDTO.UpcomingBlockHeight {
			status = constants.Processing
		}

		// Update the order status and keep the wallet associated with the order.
		return true, listener.updatePaymentOrderStatus(order, status, totalTransferred.String(), blockHeight)
	}

	return false, nil
}

func (listener *tokenTransferListener) updatePaymentOrderStatus(
	order dto.PaymentOrderDTO,
	status, transferredAmount string,
	blockHeight uint64,
) error {
	// Convert transferredAmount from Wei to Native token
	transferredAmountInEth, err := utils.ConvertSmallestUnitToFloatToken(transferredAmount, listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("failed to convert transferred amount: %w", err)
	}

	// Update order fields
	order.Transferred = transferredAmountInEth
	order.Status = status

	// Ensure order still exists before updating the queue
	newIndex, stillExists := listener.queue.GetIndex(order.PaymentAddress + "_" + order.Symbol)
	if !stillExists {
		logger.GetLogger().Warnf("Order ID %d was removed from the queue before it could be updated", order.ID)
		return fmt.Errorf("order no longer exists in the queue")
	}

	// Update item in queue
	if err := listener.queue.ReplaceItemAtIndex(newIndex, order); err != nil {
		return fmt.Errorf("failed to update order in queue: %w", err)
	}

	// Update the payment order in database
	err = listener.paymentOrderUCase.UpdatePaymentOrder(
		listener.ctx,
		order.ID,
		&blockHeight,
		nil,
		&status,
		&transferredAmountInEth,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment order in DB: %w", err)
	}

	logger.GetLogger().Infof("Successfully updated order ID %d to status '%s' with transferred amount: %s on block %d",
		order.ID, status, transferredAmountInEth, blockHeight)

	return nil
}

// dequeueOrders removes expired or successful orders from the queue.
func (listener *tokenTransferListener) dequeueOrders() {
	// Retrieve all current orders in the queue
	orders := listener.queue.GetItems()

	var dequeueExpiredOrders []dto.PaymentOrderDTO
	// Iterate over current orders in the queue
	for index, order := range orders {
		// Double-check the order network before dequeuing
		if order.Network != listener.network.String() {
			continue
		}
		// Check if the order needs to be dequeued
		if listener.shouldDequeueOrder(order) {
			// Update the order status to 'Expired' if it has expired
			if order.Status != constants.Success {
				orders[index].Status = constants.Expired
				dequeueExpiredOrders = append(dequeueExpiredOrders, orders[index])
			}

			// Remove the order from the queue
			if err := listener.queue.Dequeue(func(o dto.PaymentOrderDTO) bool {
				return o.ID == order.ID
			}); err != nil {
				logger.GetLogger().Errorf("Failed to dequeue order ID: %d, error: %v", order.ID, err)
				return
			}
		}
	}

	// Update the statuses of dequeued expired orders
	if len(dequeueExpiredOrders) > 0 {
		// Send webhooks for expired orders
		listener.sendWebhookForOrders(dequeueExpiredOrders, constants.Expired)

		// Collect order IDs
		var orderIDs []uint64
		for _, order := range dequeueExpiredOrders {
			orderIDs = append(orderIDs, order.ID)
		}

		// Update the statuses of the dequeued expired orders
		err := listener.paymentOrderUCase.BatchUpdateOrdersToExpired(listener.ctx, orderIDs)
		// Log the result of the batch update
		if err != nil {
			// Log all failed order IDs in case of an error
			var failedOrderIDs []uint64
			for _, order := range dequeueExpiredOrders {
				failedOrderIDs = append(failedOrderIDs, order.ID)
			}
			logger.GetLogger().Errorf("Failed to update expired orders with IDs: %v, error: %v", failedOrderIDs, err)
		} else {
			// Log all successful order IDs
			var successfulOrderIDs []uint64
			for _, order := range dequeueExpiredOrders {
				successfulOrderIDs = append(successfulOrderIDs, order.ID)
			}
			logger.GetLogger().Infof("Successfully updated orders with IDs: %v to expired status", successfulOrderIDs)
		}
	}
}

// shouldDequeueOrder checks if an order is expired or succeeded.
func (listener *tokenTransferListener) shouldDequeueOrder(order dto.PaymentOrderDTO) bool {
	if order.Status == constants.Processing {
		return false
	}
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

// sendWebhookForOrders sends webhooks for the given orders with the specified status.
func (listener *tokenTransferListener) sendWebhookForOrders(orders []dto.PaymentOrderDTO, status string) {
	if len(orders) == 0 {
		logger.GetLogger().Info("No orders to send webhooks for.")
		return
	}

	// Prepare the payment order DTOs for the webhook
	var orderDTOs []dto.PaymentOrderDTOResponse
	for _, order := range orders {
		paymentOrderDTO := dto.PaymentOrderDTOResponse{
			ID:             order.ID,
			RequestID:      order.RequestID,
			Network:        order.Network,
			Amount:         order.Amount,
			Transferred:    order.Transferred,
			Status:         status,
			WebhookURL:     order.WebhookURL,
			Symbol:         order.Symbol,
			PaymentAddress: order.PaymentAddress,
			Expired:        uint64(order.ExpiredTime.Unix()),
		}
		orderDTOs = append(orderDTOs, paymentOrderDTO)
	}

	// Send webhooks
	errors := utils.SendWebhooks(
		listener.ctx,
		utils.ToInterfaceSlice(orderDTOs),
		func(order interface{}) string {
			return order.(dto.PaymentOrderDTOResponse).WebhookURL
		},
	)
	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	} else {
		logger.GetLogger().Infof("All webhooks for %s orders sent successfully.", status)
	}
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
