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
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	settypes "github.com/genefriendway/onchain-handler/internal/adapters/orderset/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	listenertypes "github.com/genefriendway/onchain-handler/internal/listeners/types"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// tokenTransferListener listens for token transfers and processes them using a set of payment orders.
type tokenTransferListener struct {
	ctx                      context.Context
	cacheRepo                cachetypes.CacheRepository
	baseEventListener        listenertypes.BaseEventListener
	paymentOrderUCase        ucasetypes.PaymentOrderUCase
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase
	paymentStatisticsUCase   ucasetypes.PaymentStatisticsUCase
	paymentWalletUCase       ucasetypes.PaymentWalletUCase
	network                  constants.NetworkType
	tokenContractAddress     string
	tokenDecimals            uint8
	parsedABI                abi.ABI
	orderSet                 settypes.Set[dto.PaymentOrderDTO]
	mu                       sync.Mutex // Mutex for ticker synchronization
}

// NewTokenTransferListener creates a new tokenTransferListener with a payment order set.
func NewTokenTransferListener(
	ctx context.Context,
	cacheRepo cachetypes.CacheRepository,
	baseEventListener listenertypes.BaseEventListener,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	network constants.NetworkType,
	tokenContractAddress string,
	orderSet settypes.Set[dto.PaymentOrderDTO],
) (listenertypes.EventListener, error) {
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
		orderSet:                 orderSet,
		parsedABI:                parsedABI,
	}

	// Init the order set
	if err := listener.orderSet.Fill(
		paymentOrderUCase.GetActivePaymentOrders,
	); err != nil {
		logger.GetLogger().Errorf("Failed to init the order set %s: %v", listener.network.String(), err)
	}

	go listener.startCleanSetTicker(constants.CleanSetInterval)

	return listener, nil
}

// startCleanSetTicker starts a ticker that triggers cleaning expired or successful orders at a specified interval.
func (listener *tokenTransferListener) startCleanSetTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to lock before cleaning, ensuring no overlap
			listener.mu.Lock()
			logger.GetLogger().Debug("Starting cleaning operation...")
			listener.removeOrders()
			logger.GetLogger().Debug("Clean set operation finished.")
			listener.mu.Unlock()
		case <-listener.ctx.Done():
			logger.GetLogger().Debugf("Stopping clean set ticker: %v", listener.ctx.Err())
			return
		}
	}
}

func (listener *tokenTransferListener) fetchOrderDetailsFromSet(
	key string, transferEvent blockchain.TransferEvent, tokenSymbol string,
) (*dto.PaymentOrderDTO, error) {
	// Fetch order from the set
	order, exists := listener.orderSet.GetItem(key)
	if !exists {
		logger.GetLogger().Infof("No matching order found in set for address %s and token %s", transferEvent.To.Hex(), tokenSymbol)
		return nil, nil // Return nil instead of an error if no order is found
	}

	// Ensure order matches the correct network
	if order.Network != listener.network.String() {
		logger.GetLogger().Infof("Skipping order ID %d as it belongs to a different network: %s", order.ID, order.Network)
		return nil, nil
	}

	return &order, nil
}

func (listener *tokenTransferListener) parseAndProcessRealtimeTransferEvent(vLog types.Log) (any, error) {
	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := conf.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil,
			fmt.Errorf(
				"failed to unpack realtime transfer event on network %s for address %s, block number %d: %w",
				listener.network.String(),
				vLog.Address.Hex(),
				vLog.BlockNumber,
				err,
			)
	}

	// Create a unique key for the order
	key := transferEvent.To.Hex() + "_" + tokenSymbol

	// Fetch the order details from the set
	order, err := listener.fetchOrderDetailsFromSet(key, transferEvent, tokenSymbol)
	if err != nil || order == nil {
		return nil, err
	}
	logger.GetLogger().Infof("Found order ID %d in set: %v", order.ID, order)

	// Get block number from the event
	upcomingBlockHeight := vLog.BlockNumber

	// Prevent unnecessary status update
	if order.Status == constants.Success {
		logger.GetLogger().Infof("Skipping order ID %d as it is already in SUCCESS status", order.ID)
		return nil, nil
	}
	if order.BlockHeight >= upcomingBlockHeight || order.UpcomingBlockHeight >= upcomingBlockHeight {
		logger.GetLogger().Infof(
			"Skipping event from older block %d for order ID %d (current block height: %d, current upcoming block height: %d)",
			upcomingBlockHeight,
			order.ID,
			order.BlockHeight,
			order.UpcomingBlockHeight,
		)
		return nil, nil
	}

	// Update the order status to 'Processing'
	status := constants.Processing
	err = listener.paymentOrderUCase.UpdatePaymentOrder(listener.ctx, order.ID, nil, &upcomingBlockHeight, &status, nil, nil)
	if err != nil {
		logger.GetLogger().Errorf("Failed to update order status to processing on network %s for order ID %d, error: %v", listener.network.String(), order.ID, err)
		return nil, err
	}
	logger.GetLogger().Infof(
		"Updated order ID %d to status 'Processing' on network %s, block height: %d",
		order.ID,
		listener.network.String(),
		upcomingBlockHeight,
	)

	// Update the order status in the set
	order.Status = status
	order.UpcomingBlockHeight = upcomingBlockHeight

	if err := listener.orderSet.UpdateItem(key, *order); err != nil {
		logger.GetLogger().Errorf("Failed to update order in set: %v", err)
		return nil, err
	}

	return transferEvent, nil
}

// parseAndProcessConfirmedTransferEvent parses and processes a confirmed transfer event, checking if it matches any payment order in the set.
func (listener *tokenTransferListener) parseAndProcessConfirmedTransferEvent(vLog types.Log) (any, error) {
	// Retrieve the token symbol for the event's contract address
	tokenSymbol, err := conf.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return nil,
			fmt.Errorf("failed to get token symbol from token contract address %s: %w", vLog.Address.Hex(), err)
	}

	// Unpack the transfer event
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, listener.parsedABI)
	if err != nil {
		return nil,
			fmt.Errorf(
				"failed to unpack confirmed transfer event on network %s for address %s, block number %d: %w",
				listener.network.String(),
				vLog.Address.Hex(),
				vLog.BlockNumber,
				err,
			)
	}

	// Create a unique key for the order
	key := transferEvent.To.Hex() + "_" + tokenSymbol

	// Fetch the order details from the set
	order, err := listener.fetchOrderDetailsFromSet(key, transferEvent, tokenSymbol)
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

	// Recheck if the processed order is already SUCCESS
	if err := listener.recheckOrder(&processedOrder); err != nil {
		logger.GetLogger().Errorf("Recheck and release wallet failed for order ID %d: %v", processedOrder.ID, err)
	}

	// Ready to send webhook for the processed order
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

	// Get newest order state in cache or DB
	orderDTO, err := listener.paymentOrderUCase.GetPaymentOrderByID(listener.ctx, order.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get order by ID %d, error: %v", order.ID, err)
		return false, err
	}

	// Calculate total transferred amount from EventHistories
	totalTransferred := big.NewInt(0)
	for _, event := range orderDTO.EventHistories {
		amountWei, err := utils.ConvertFloatTokenToSmallestUnit(event.Amount, listener.tokenDecimals)
		if err != nil {
			logger.GetLogger().Warnf("Failed to convert event amount to Wei, tx: %s, err: %v", event.TransactionHash, err)
			continue
		}
		totalTransferred = new(big.Int).Add(totalTransferred, amountWei)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred = new(big.Int).Add(totalTransferred, transferEvent.Value)

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		logger.GetLogger().Infof("Processed full payment on network %s for order ID: %d", listener.network.String(), order.ID)

		status := constants.Success
		// Check if order is still 'Processing' or needs to be marked as 'Success'.
		/*
			if blockHeight < order.UpcomingBlockHeight {
				status = constants.Processing
			}
		*/

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

	// Update item in set
	key := order.PaymentAddress + "_" + order.Symbol
	if err := listener.orderSet.UpdateItem(key, order); err != nil {
		return fmt.Errorf("failed to update order in set: %w", err)
	}

	logger.GetLogger().Infof("Successfully updated order ID %d to status '%s' with transferred amount: %s on block %d",
		order.ID, status, transferredAmountInEth, blockHeight)

	return nil
}

// removeOrders removes expired or successful orders from the set.
func (listener *tokenTransferListener) removeOrders() {
	// Retrieve all current orders in the set
	orders := listener.orderSet.GetAll()

	var cleanedExpiredOrders []dto.PaymentOrderDTO
	// Iterate over current orders in the set
	for index, order := range orders {
		// Double-check the order network
		if order.Network != listener.network.String() {
			continue
		}
		// Check if the order needs to be cleaned
		if listener.shouldCleanOrder(order) {
			// Update the order status to 'Expired' if it has expired
			if order.Status != constants.Success {
				orders[index].Status = constants.Expired
				cleanedExpiredOrders = append(cleanedExpiredOrders, orders[index])
			}

			// Remove the order from the set
			if !listener.orderSet.Remove(func(o dto.PaymentOrderDTO) bool {
				return o.ID == order.ID
			}) {
				logger.GetLogger().Errorf("Failed to remove order ID from the order set: %d", order.ID)
				return
			}

		}
	}

	// Update the statuses of cleaned expired orders
	if len(cleanedExpiredOrders) > 0 {
		// Send webhooks for expired orders
		listener.sendWebhookForOrders(cleanedExpiredOrders, constants.Expired)

		// Collect order IDs
		var orderIDs []uint64
		for _, order := range cleanedExpiredOrders {
			orderIDs = append(orderIDs, order.ID)
		}

		// Update the statuses of the cleaned expired orders
		err := listener.paymentOrderUCase.BatchUpdateOrdersToExpired(listener.ctx, orderIDs)
		// Log the result of the batch update
		if err != nil {
			// Log all failed order IDs in case of an error
			var failedOrderIDs []uint64
			for _, order := range cleanedExpiredOrders {
				failedOrderIDs = append(failedOrderIDs, order.ID)
			}
			logger.GetLogger().Errorf("Failed to update expired orders with IDs: %v, error: %v", failedOrderIDs, err)
		} else {
			// Log all successful order IDs
			var successfulOrderIDs []uint64
			for _, order := range cleanedExpiredOrders {
				successfulOrderIDs = append(successfulOrderIDs, order.ID)
			}
			logger.GetLogger().Infof("Successfully updated orders with IDs: %v to expired status", successfulOrderIDs)
		}
	}
}

// shouldCleanOrder checks if an order is expired or succeeded.
func (listener *tokenTransferListener) shouldCleanOrder(order dto.PaymentOrderDTO) bool {
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
		func(order any) string {
			return order.(dto.PaymentOrderDTOResponse).WebhookURL
		},
	)
	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	} else {
		logger.GetLogger().Infof("All webhooks for %s orders sent successfully.", status)
	}
}

func (listener *tokenTransferListener) recheckOrder(processedOrder *dto.PaymentOrderDTOResponse) error {
	// Already success, no further processing required.
	if processedOrder.Status == constants.Success {
		logger.GetLogger().Infof("Order ID %d already SUCCESS and wallet released.", processedOrder.ID)
		return nil
	}

	// Convert order amount to smallest unit (Wei)
	orderAmountWei, err := utils.ConvertFloatTokenToSmallestUnit(processedOrder.Amount, listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("failed to convert order amount: %w", err)
	}

	// Calculate total transferred amount from PaymentHistory
	totalTransferredWei := big.NewInt(0)
	for _, event := range processedOrder.EventHistories {
		eventAmountWei, err := utils.ConvertFloatTokenToSmallestUnit(event.Amount, listener.tokenDecimals)
		if err != nil {
			return fmt.Errorf("failed to convert event amount (tx: %s): %w", event.TransactionHash, err)
		}
		totalTransferredWei.Add(totalTransferredWei, eventAmountWei)
	}

	minimumAcceptedAmount := payment.CalculatePaymentCoveringAsDiscount(
		orderAmountWei, conf.GetPaymentCovering(), listener.tokenDecimals,
	)

	// Check if total transferred amount is sufficient
	if totalTransferredWei.Cmp(minimumAcceptedAmount) < 0 {
		logger.GetLogger().Infof(
			"Order ID %d has insufficient amount transferred (%s Wei) for SUCCESS status (minimum required: %s Wei).",
			processedOrder.ID,
			totalTransferredWei.String(),
			minimumAcceptedAmount.String(),
		)
		return nil
	}

	// Update DB status and release wallet
	if err := listener.paymentOrderUCase.UpdateOrderToSuccessAndReleaseWallet(listener.ctx, processedOrder.ID); err != nil {
		return fmt.Errorf("failed to update order status to SUCCESS and release wallet: %w", err)
	}

	logger.GetLogger().Infof("Successfully updated order ID %d to SUCCESS status independently.", processedOrder.ID)

	// Update processed order status
	processedOrder.Status = constants.Success
	processedOrder.Transferred, err = utils.ConvertSmallestUnitToFloatToken(totalTransferredWei.String(), listener.tokenDecimals)
	if err != nil {
		return fmt.Errorf("failed to convert total transferred amount back to float: %w", err)
	}

	// Update the order in the set
	key := processedOrder.PaymentAddress + "_" + processedOrder.Symbol
	orderInSet, exists := listener.orderSet.GetItem(key)
	if !exists {
		logger.GetLogger().Warnf("Order ID %d not found in set for key %s during recheck.", processedOrder.ID, key)
		return nil
	}

	orderInSet.Status = constants.Success
	orderInSet.Transferred = processedOrder.Transferred
	if err := listener.orderSet.UpdateItem(key, orderInSet); err != nil {
		return fmt.Errorf("failed to update order in set: %w", err)
	}

	return nil
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
