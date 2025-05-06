package workers

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	workertypes "github.com/genefriendway/onchain-handler/internal/workers/types"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// expiredOrderCatchupWorker is a worker that processes expired orders.
type expiredOrderCatchupWorker struct {
	paymentOrderUCase        ucasetypes.PaymentOrderUCase
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase
	paymentStatisticsUCase   ucasetypes.PaymentStatisticsUCase
	paymentWalletUCase       ucasetypes.PaymentWalletUCase
	blockStateUCase          ucasetypes.BlockStateUCase
	cacheRepo                cachetypes.CacheRepository
	tokenContractAddresses   []string
	tokenDecimalsMap         map[string]uint8
	parsedABI                abi.ABI
	ethClient                clienttypes.Client
	network                  constants.NetworkType
	confirmationDepth        uint64
	processedOrderIDs        map[uint64]struct{}
	isRunning                bool       // Tracks if catchup is running
	mu                       sync.Mutex // Mutex to protect the isRunning flag
}

func NewExpiredOrderCatchupWorker(
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	blockStateUCase ucasetypes.BlockStateUCase,
	cacheRepo cachetypes.CacheRepository,
	tokenContractAddresses []string,
	ethClient clienttypes.Client,
	network constants.NetworkType,
) workertypes.Worker {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		logger.GetLogger().Infof("failed to parse ERC20 ABI: %v", err)
		return nil
	}

	tokenDecimalsMap := make(map[string]uint8)
	for _, addr := range tokenContractAddresses {
		decimals, err := blockchain.GetTokenDecimalsFromCache(addr, network.String(), cacheRepo)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get token decimals from cache for %s: %v", addr, err)
			return nil
		}
		tokenDecimalsMap[addr] = decimals
	}

	confirmationDepth, err := blockchain.GetConfirmationDepth(network)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get confirmation depth for network %s: %v", network.String(), err)
		confirmationDepth = constants.DefaultConfirmationDepth // Fallback to a default value
	}

	return &expiredOrderCatchupWorker{
		paymentOrderUCase:        paymentOrderUCase,
		paymentEventHistoryUCase: paymentEventHistoryUCase,
		paymentStatisticsUCase:   paymentStatisticsUCase,
		paymentWalletUCase:       paymentWalletUCase,
		blockStateUCase:          blockStateUCase,
		cacheRepo:                cacheRepo,
		tokenContractAddresses:   tokenContractAddresses,
		tokenDecimalsMap:         tokenDecimalsMap,
		parsedABI:                parsedABI,
		ethClient:                ethClient,
		network:                  network,
		confirmationDepth:        confirmationDepth,
		processedOrderIDs:        make(map[uint64]struct{}),
	}
}

func (w *expiredOrderCatchupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.ExpiredOrderCatchupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx) // Run the catchup process in a separate goroutine
		case <-ctx.Done():
			logger.GetLogger().Infof("Shutting down expiredOrderCatchupWorker on network %s", w.network.String())
			return
		}
	}
}

// ensures that expiredOrderCatchupWorker doesn't overlap by checking the isRunning flag.
func (w *expiredOrderCatchupWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warnf("Previous catchupExpiredOrders on network %s run still in progress, skipping this cycle", w.network.String())
		w.mu.Unlock()
		return
	}

	// Mark as running
	w.isRunning = true
	w.mu.Unlock()

	// Perform the catch-up process
	w.catchupExpiredOrders(ctx)

	// Mark as not running
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()
}

func (w *expiredOrderCatchupWorker) catchupExpiredOrders(ctx context.Context) {
	// Fetch expired orders from the repository
	expiredOrders, err := w.paymentOrderUCase.GetExpiredPaymentOrders(ctx, w.network)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve expired payment orders on network %s: %v", w.network.String(), err)
		return
	}

	// No expired orders to process
	if len(expiredOrders) == 0 {
		logger.GetLogger().Infof("No expired orders on network %s found", w.network.String())
		return
	}

	// Find the smallest block height from the expired orders
	minBlockHeight := expiredOrders[0].BlockHeight

	// Define the maximum block we should query up to
	latestBlock, err := w.blockStateUCase.GetLatestBlock(ctx, w.network)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve latest block number on network %s: %v", w.network.String(), err)
		return
	}

	// Calculate the effective latest block considering the confirmation depth.
	effectiveLatestBlock := latestBlock - w.confirmationDepth

	// Start processing logs from the smallest block height to the effective latest block
	w.processExpiredOrders(ctx, minBlockHeight, effectiveLatestBlock, expiredOrders)

	// Update processed block height for non processed orders
	var nonProcessedOrders []dto.PaymentOrderDTO
	for index, expiredOrder := range expiredOrders {
		if _, exists := w.processedOrderIDs[expiredOrder.ID]; !exists {
			expiredOrders[index].BlockHeight = effectiveLatestBlock
			nonProcessedOrders = append(nonProcessedOrders, expiredOrders[index])
		}
	}

	if len(nonProcessedOrders) > 0 {
		// Attempt to batch update the block heights
		err = w.paymentOrderUCase.BatchUpdateOrderBlockHeights(ctx, nonProcessedOrders)
		if err != nil {
			// Log all failed order IDs
			var failedOrderIDs []uint64
			for _, order := range nonProcessedOrders {
				failedOrderIDs = append(failedOrderIDs, order.ID)
			}
			logger.GetLogger().Errorf("Failed to update block heights for orders with IDs: %v, error: %v", failedOrderIDs, err)
		} else {
			// Log all successful order IDs
			var successfulOrderIDs []uint64
			for _, order := range nonProcessedOrders {
				successfulOrderIDs = append(successfulOrderIDs, order.ID)
			}
			logger.GetLogger().Infof("Successfully updated block heights for orders with IDs: %v", successfulOrderIDs)
			w.processedOrderIDs = make(map[uint64]struct{})
		}
	}
}

// processExpiredOrders processes logs from the blockchain starting from the given block height
func (w *expiredOrderCatchupWorker) processExpiredOrders(ctx context.Context, startBlock, endBlock uint64, expiredOrders []dto.PaymentOrderDTO) {
	logger.GetLogger().Infof("Processing expired orders on network %s starting from block %d to block %d", w.network.String(), startBlock, endBlock)

	// Check for invalid end block height
	if endBlock <= 0 {
		logger.GetLogger().Warnf("End block (%d) is non-positive. Skipping processing for expired orders on network %s", endBlock, w.network.String())
		return
	}

	// Safeguard if start block is beyond the end block
	if startBlock > endBlock {
		logger.GetLogger().Warnf("Start block %d is beyond the end block %d. No logs to process on network %s", startBlock, endBlock, w.network.String())
		return
	}

	// Process logs in chunks of DefaultBlockOffset
	var addresses []common.Address
	for _, tokenAddress := range w.tokenContractAddresses {
		addresses = append(addresses, common.HexToAddress(tokenAddress))
	}
	for chunkStart := startBlock; chunkStart <= endBlock; chunkStart += constants.DefaultBlockOffset {
		chunkEnd := min(chunkStart+constants.DefaultBlockOffset-1, endBlock)

		logger.GetLogger().Debugf("Expired Order Catchup Worker: Processing block chunk from %d to %d on network %s", chunkStart, chunkEnd, w.network.String())

		// Poll logs from blockchain for this block range
		logs, err := w.ethClient.PollForLogsFromBlock(ctx, addresses, chunkStart, chunkEnd)
		if err != nil {
			logger.GetLogger().Errorf("Failed to poll logs on network %s from block range %d-%d: %v", w.network.String(), chunkStart, chunkEnd, err)
			continue
		}

		// Process each log entry and match with expired orders
		for _, logEntry := range logs {
			err := w.processLog(ctx, logEntry, expiredOrders, logEntry.BlockNumber)
			if err != nil {
				logger.GetLogger().Errorf("Error processing log entry on network %s: %v", w.network.String(), err)
				continue
			}
		}
	}
}

// processLog processes a single log entry from the blockchain
func (w *expiredOrderCatchupWorker) processLog(
	ctx context.Context,
	vLog types.Log,
	orders []dto.PaymentOrderDTO,
	blockHeight uint64,
) error {
	logger.GetLogger().Infof("Processing log entry on network %s from address: %s", w.network.String(), vLog.Address.Hex())

	tokenSymbol, err := conf.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return fmt.Errorf("failed to get token symbol from token contract address on network %s: %w", w.network.String(), err)
	}

	tokenDecimals := w.tokenDecimalsMap[vLog.Address.Hex()]

	// Unpack the transfer event from the log
	transferEvent, err := blockchain.UnpackTransferEvent(vLog, w.parsedABI)
	if err != nil {
		return fmt.Errorf("failed to unpack transfer event on network %s: %w", w.network.String(), err)
	}

	// Iterate over all expired orders to find a matching wallet address
	for index, order := range orders {
		// Check if the order matches the transfer event based on the wallet address and token symbol
		if !w.isMatchingOrder(order, transferEvent, tokenSymbol) {
			continue
		}
		// Found a matching order, now process the payment for that order
		logger.GetLogger().Infof("Matched transfer to wallet %s for order ID on network %s: %d", transferEvent.To.Hex(), w.network.String(), order.ID)

		// Call processOrderPayment to handle the order update logic based on the transfer event
		isUpdated, err := w.processOrderPayment(ctx, &orders[index], transferEvent, blockHeight, tokenDecimals)
		if err != nil {
			return fmt.Errorf("failed to process order payment for order ID %d on network %s: %w", order.ID, w.network.String(), err)
		}

		if !isUpdated {
			continue
		}

		// Convert the transfer event value to ETH
		transferEventValueInEth, err := utils.ConvertSmallestUnitToFloatToken(transferEvent.Value.String(), tokenDecimals)
		if err != nil {
			logger.GetLogger().Errorf("Failed to convert transfer event value to ETH for order ID %d: %v", order.ID, err)
			return err
		}

		// Create payment event history for the order
		if err := w.createPaymentEventHistory(ctx, order, transferEventValueInEth, transferEvent, tokenSymbol, vLog.Address.Hex(), vLog.TxHash.Hex()); err != nil {
			return fmt.Errorf("failed to create payment event history for order ID %d on network %s: %w", order.ID, w.network.String(), err)
		}
		logger.GetLogger().Infof("Successfully processed order ID: %d on network %s with transferred amount: %s", order.ID, w.network.String(), transferEvent.Value.String())

		// Get the payment order by ID to send the webhook
		paymentOrderDTO, err := w.paymentOrderUCase.GetPaymentOrderByID(ctx, order.ID)
		if err != nil {
			return fmt.Errorf("failed to get payment order by ID %d on network %s: %w", order.ID, w.network.String(), err)
		}

		// Increment payment statistics
		granularity := constants.Daily
		if err = w.paymentStatisticsUCase.IncrementStatistics(
			ctx,
			granularity,
			utils.GetPeriodStart(granularity, time.Now()),
			nil,
			&transferEventValueInEth,
			order.Symbol,
			order.VendorID,
		); err != nil {
			logger.GetLogger().Errorf("Failed to increment payment statistics for order ID %d on network %s: %v", order.ID, w.network.String(), err)
		}

		// Add the transferred amount to the payment wallet balance
		if err = w.paymentWalletUCase.AddPaymentWalletBalance(ctx, order.Wallet.ID, transferEventValueInEth, w.network, order.Symbol); err != nil {
			logger.GetLogger().Errorf("Failed to add payment wallet balance on network %s for order ID %d, error: %v", w.network.String(), order.ID, err)
			continue
		}

		// Send webhook if webhook URL is present
		if paymentOrderDTO.WebhookURL != "" {
			// Use a separate goroutine for webhook sending
			go func() {
				if err := utils.SendWebhook(paymentOrderDTO, paymentOrderDTO.WebhookURL); err != nil {
					logger.GetLogger().Errorf("Failed to send webhook for order ID %d on network %s: %v", order.ID, w.network.String(), err)
				}
			}()
		}

		logger.GetLogger().Infof("Successfully processed order ID: %d on network %s with transferred amount: %s", order.ID, w.network.String(), transferEvent.Value.String())
		return nil // Stop once we've processed the matching order
	}

	// No matching order found for this log entry
	logger.GetLogger().Infof("No matching expired order found for transfer to address: %s on network %s", transferEvent.To.Hex(), w.network.String())
	return nil
}

// isMatchingOrder checks if the order matches the transfer event based on the wallet address and token symbol
func (w *expiredOrderCatchupWorker) isMatchingOrder(order dto.PaymentOrderDTO, transferEvent blockchain.TransferEvent, tokenSymbol string) bool {
	return order.Status != constants.Success &&
		strings.EqualFold(transferEvent.To.Hex(), order.Wallet.Address) &&
		strings.EqualFold(order.Symbol, tokenSymbol)
}

// createPaymentEventHistory constructs and stores the payment event history
func (w *expiredOrderCatchupWorker) createPaymentEventHistory(
	ctx context.Context,
	order dto.PaymentOrderDTO,
	transferEventValueInEth string,
	transferEvent blockchain.TransferEvent,
	tokenSymbol, contractAddress, txHash string,
) error {
	payloads := []dto.PaymentEventPayloadDTO{
		{
			PaymentOrderID:  order.ID,
			TransactionHash: txHash,
			FromAddress:     transferEvent.From.Hex(),
			ToAddress:       transferEvent.To.Hex(),
			ContractAddress: contractAddress,
			TokenSymbol:     tokenSymbol,
			Amount:          transferEventValueInEth,
			Network:         w.network.String(),
		},
	}

	return w.paymentEventHistoryUCase.CreatePaymentEventHistory(ctx, payloads)
}

// processOrderPayment handles the payment for an order based on the transfer event details.
// It updates the order status and wallet usage based on the payment amount.
func (w *expiredOrderCatchupWorker) processOrderPayment(
	ctx context.Context,
	order *dto.PaymentOrderDTO,
	transferEvent blockchain.TransferEvent,
	blockHeight uint64,
	tokenDecimals uint8,
) (bool, error) {
	if blockHeight <= order.BlockHeight {
		logger.GetLogger().Infof("Processed order: %d on network %s. Ignore this turn.", order.ID, w.network.String())
		return false, nil
	}

	// Convert order amount into the appropriate unit (e.g., wei)
	orderAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Amount, tokenDecimals)
	if err != nil {
		return false, fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := payment.CalculatePaymentCoveringAsDiscount(orderAmount, conf.GetPaymentCovering(), tokenDecimals)

	// Get newest order state in cache or DB
	orderDTO, err := w.paymentOrderUCase.GetPaymentOrderByID(ctx, order.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get order by ID %d, error: %v", order.ID, err)
		return false, err
	}

	// Calculate total transferred amount from EventHistories
	totalTransferred := big.NewInt(0)
	for _, event := range orderDTO.EventHistories {
		amountWei, err := utils.ConvertFloatTokenToSmallestUnit(event.Amount, tokenDecimals)
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
		logger.GetLogger().Infof("Processed full payment on network %s for order ID: %d", w.network.String(), order.ID)

		if order.BlockHeight < order.UpcomingBlockHeight {
			// Update the order tranferred and keep the wallet associated with the order.
			return w.updatePaymentOrderStatus(ctx, order, order.Status, totalTransferred.String(), blockHeight, tokenDecimals)
		}

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return w.updatePaymentOrderStatus(ctx, order, constants.Success, totalTransferred.String(), blockHeight, tokenDecimals)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		logger.GetLogger().Infof("Processed partial payment on network %s for order ID: %d", w.network.String(), order.ID)

		// Update the order tranferred and keep the wallet associated with the order.
		return w.updatePaymentOrderStatus(ctx, order, order.Status, totalTransferred.String(), blockHeight, tokenDecimals)
	}

	return false, nil
}

// updatePaymentOrderStatus updates the payment order with the new status and transferred amount.
func (w *expiredOrderCatchupWorker) updatePaymentOrderStatus(
	ctx context.Context,
	order *dto.PaymentOrderDTO,
	status, transferredAmount string,
	blockHeight uint64,
	tokenDecimals uint8,
) (bool, error) {
	// Convert transferredAmount from Wei to Eth (Ether)
	transferredAmountInEth, err := utils.ConvertSmallestUnitToFloatToken(transferredAmount, tokenDecimals)
	if err != nil {
		return false, fmt.Errorf("updatePaymentOrderStatus error: %v", err)
	}

	// Update item in list
	order.Transferred = transferredAmountInEth
	order.Status = status

	// Save the updated order to the repository
	err = w.paymentOrderUCase.UpdatePaymentOrder(ctx, order.ID, &blockHeight, nil, &status, &transferredAmountInEth, nil)
	if err != nil {
		return false, fmt.Errorf("failed to update payment order status for order ID %d: %w", order.ID, err)
	}

	logger.GetLogger().Infof("Order ID %d on network %s updated to status %s with transferred amount: %s", order.ID, w.network.String(), status, transferredAmount)
	w.processedOrderIDs[order.ID] = struct{}{}
	return true, nil
}
