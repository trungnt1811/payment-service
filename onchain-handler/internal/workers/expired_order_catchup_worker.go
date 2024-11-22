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
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// expiredOrderCatchupWorker is a worker that processes expired orders.
type expiredOrderCatchupWorker struct {
	config                   *conf.Configuration
	paymentOrderUCase        interfaces.PaymentOrderUCase
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase
	cacheRepo                infrainterfaces.CacheRepository
	contractAddress          string
	parsedABI                abi.ABI
	ethClient                pkginterfaces.Client
	network                  constants.NetworkType
	processedOrderIDs        map[uint64]struct{}
	isRunning                bool       // Tracks if catchup is running
	mu                       sync.Mutex // Mutex to protect the isRunning flag
}

func NewExpiredOrderCatchupWorker(
	config *conf.Configuration,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	cacheRepo infrainterfaces.CacheRepository,
	contractAddress string,
	ethClient pkginterfaces.Client,
	network constants.NetworkType,
) interfaces.Worker {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		logger.GetLogger().Infof("failed to parse ERC20 ABI: %v", err)
		return nil
	}
	return &expiredOrderCatchupWorker{
		config:                   config,
		paymentOrderUCase:        paymentOrderUCase,
		paymentEventHistoryUCase: paymentEventHistoryUCase,
		cacheRepo:                cacheRepo,
		contractAddress:          contractAddress,
		parsedABI:                parsedABI,
		ethClient:                ethClient,
		network:                  network,
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
			logger.GetLogger().Infof("Shutting down expiredOrderCatchupWorker on network %s", string(w.network))
			return
		}
	}
}

// ensures that expiredOrderCatchupWorker doesn't overlap by checking the isRunning flag.
func (w *expiredOrderCatchupWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warnf("Previous catchupExpiredOrders on network %s run still in progress, skipping this cycle", string(w.network))
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
		logger.GetLogger().Errorf("Failed to retrieve expired payment orders on network %s: %v", string(w.network), err)
		return
	}

	// No expired orders to process
	if len(expiredOrders) == 0 {
		logger.GetLogger().Infof("No expired orders on network %s found", string(w.network))
		return
	}

	// Find the smallest block height from the expired orders
	minBlockHeight := expiredOrders[0].BlockHeight

	// Define the maximum block we should query up to
	latestBlock, err := blockchain.GetLatestBlockFromCacheOrBlockchain(
		ctx,
		string(w.network),
		w.cacheRepo,
		w.ethClient,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve latest block number on network %s: %v", string(w.network), err)
		return
	}

	// Calculate the effective latest block considering the confirmation depth.
	effectiveLatestBlock := latestBlock - constants.ConfirmationDepth

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
	logger.GetLogger().Infof("Processing expired orders on network %s starting from block %d to block %d", string(w.network), startBlock, endBlock)

	// Check for invalid end block height
	if endBlock <= 0 {
		logger.GetLogger().Warnf("End block (%d) is non-positive. Skipping processing for expired orders on network %s", endBlock, string(w.network))
		return
	}

	// Safeguard if start block is beyond the end block
	if startBlock > endBlock {
		logger.GetLogger().Warnf("Start block %d is beyond the end block %d. No logs to process on network %s", startBlock, endBlock, string(w.network))
		return
	}

	// Process logs in chunks of DefaultBlockOffset
	address := common.HexToAddress(w.contractAddress)
	for chunkStart := startBlock; chunkStart <= endBlock; chunkStart += constants.DefaultBlockOffset {
		chunkEnd := chunkStart + constants.DefaultBlockOffset - 1
		if chunkEnd > endBlock {
			chunkEnd = endBlock
		}

		logger.GetLogger().Debugf("Expired Order Catchup Worker: Processing block chunk from %d to %d on network %s", chunkStart, chunkEnd, string(w.network))

		// Poll logs from blockchain for this block range
		logs, err := w.ethClient.PollForLogsFromBlock(ctx, []common.Address{address}, chunkStart, chunkEnd)
		if err != nil {
			logger.GetLogger().Errorf("Failed to poll logs on network %s from block range %d-%d: %v", string(w.network), chunkStart, chunkEnd, err)
			continue
		}

		// Process each log entry and match with expired orders
		for _, logEntry := range logs {
			err := w.processLog(ctx, logEntry, expiredOrders, logEntry.BlockNumber)
			if err != nil {
				logger.GetLogger().Errorf("Error processing log entry on network %s: %v", string(w.network), err)
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
	logger.GetLogger().Infof("Processing log entry on network %s from address: %s", string(w.network), vLog.Address.Hex())

	tokenSymbol, err := w.config.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return fmt.Errorf("failed to get token symbol from token contract address on network %s: %w", string(w.network), err)
	}

	// Unpack the transfer event from the log
	transferEvent, err := w.unpackTransferEvent(vLog)
	if err != nil {
		return fmt.Errorf("failed to unpack transfer event on network %s: %w", string(w.network), err)
	}

	// Iterate over all expired orders to find a matching wallet address
	for index, order := range orders {
		// Check if the event's "To" address matches the order's wallet address
		if order.Status != constants.Success && strings.EqualFold(transferEvent.To.Hex(), order.Wallet.Address) && strings.EqualFold(order.Symbol, tokenSymbol) {
			// Found a matching order, now process the payment for that order
			logger.GetLogger().Infof("Matched transfer to wallet %s for order ID on network %s: %d", transferEvent.To.Hex(), string(w.network), order.ID)

			// Call processOrderPayment to handle the order update logic based on the transfer event
			isUpdated, err := w.processOrderPayment(ctx, &orders[index], transferEvent, blockHeight)
			if err != nil {
				return fmt.Errorf("failed to process order payment for order ID %d on network %s: %w", order.ID, string(w.network), err)
			}

			if isUpdated {
				if err := w.createPaymentEventHistory(ctx, order, transferEvent, tokenSymbol, vLog.Address.Hex(), vLog.TxHash.Hex()); err != nil {
					return fmt.Errorf("failed to create payment event history for order ID %d on network %s: %w", order.ID, string(w.network), err)
				}
				logger.GetLogger().Infof("Successfully processed order ID: %d on network %s with transferred amount: %s", order.ID, string(w.network), transferEvent.Value.String())
			}

			logger.GetLogger().Infof("Successfully processed order ID: %d on network %s with transferred amount: %s", order.ID, string(w.network), transferEvent.Value.String())
			return nil // Stop once we've processed the matching order
		}
	}

	// No matching order found for this log entry
	logger.GetLogger().Warnf("No matching expired order found for transfer to address: %s on network %s", transferEvent.To.Hex(), string(w.network))
	return nil
}

// createPaymentEventHistory constructs and stores the payment event history
func (w *expiredOrderCatchupWorker) createPaymentEventHistory(
	ctx context.Context,
	order dto.PaymentOrderDTO,
	transferEvent dto.TransferEventDTO,
	tokenSymbol, contractAddress, txHash string,
) error {
	transferEventValueInEth, err := utils.ConvertSmallestUnitToFloatToken(transferEvent.Value.String(), constants.NativeTokenDecimalPlaces)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert transfer event value to ETH for order ID %d: %v", order.ID, err)
		return err
	}

	payloads := []dto.PaymentEventPayloadDTO{
		{
			PaymentOrderID:  order.ID,
			TransactionHash: txHash,
			FromAddress:     transferEvent.From.Hex(),
			ToAddress:       transferEvent.To.Hex(),
			ContractAddress: contractAddress,
			TokenSymbol:     tokenSymbol,
			Amount:          transferEventValueInEth,
			Network:         string(w.network),
		},
	}

	return w.paymentEventHistoryUCase.CreatePaymentEventHistory(ctx, payloads)
}

// processOrderPayment handles the payment for an order based on the transfer event details.
// It updates the order status and wallet usage based on the payment amount.
func (w *expiredOrderCatchupWorker) processOrderPayment(
	ctx context.Context,
	order *dto.PaymentOrderDTO,
	transferEvent dto.TransferEventDTO,
	blockHeight uint64,
) (bool, error) {
	if blockHeight <= order.BlockHeight {
		logger.GetLogger().Infof("Processed order: %d on network %s. Ignore this turn.", order.ID, string(w.network))
		return false, nil
	}
	// Convert order amount and transferred amount into the appropriate unit (e.g., wei)
	orderAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Amount, constants.NativeTokenDecimalPlaces)
	if err != nil {
		return false, fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := payment.CalculatePaymentCovering(orderAmount, w.config.GetPaymentCovering())

	transferredAmount, err := utils.ConvertFloatTokenToSmallestUnit(order.Transferred, constants.NativeTokenDecimalPlaces)
	if err != nil {
		return false, fmt.Errorf("failed to convert transferred amount to wei: %v", err)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred := new(big.Int).Add(transferredAmount, transferEvent.Value)

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		logger.GetLogger().Infof("Processed full payment on network %s for order ID: %d", string(w.network), order.ID)

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return true, w.updatePaymentOrderStatus(ctx, order, constants.Success, totalTransferred.String(), blockHeight)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		logger.GetLogger().Infof("Processed partial payment on network %s for order ID: %d", string(w.network), order.ID)

		// Update the order tranferred and keep the wallet associated with the order.
		return true, w.updatePaymentOrderStatus(ctx, order, order.Status, totalTransferred.String(), blockHeight)
	}

	return false, nil
}

// updatePaymentOrderStatus updates the payment order with the new status and transferred amount.
func (w *expiredOrderCatchupWorker) updatePaymentOrderStatus(
	ctx context.Context,
	order *dto.PaymentOrderDTO,
	status, transferredAmount string,
	blockHeight uint64,
) error {
	// Convert transferredAmount from Wei to Eth (Ether)
	transferredAmountInEth, err := utils.ConvertSmallestUnitToFloatToken(transferredAmount, constants.NativeTokenDecimalPlaces)
	if err != nil {
		return fmt.Errorf("updatePaymentOrderStatus error: %v", err)
	}

	// Update item in list
	order.Transferred = transferredAmountInEth
	order.Status = status

	// Save the updated order to the repository
	err = w.paymentOrderUCase.UpdatePaymentOrder(ctx, order.ID, &blockHeight, &status, &transferredAmountInEth, nil)
	if err != nil {
		return fmt.Errorf("failed to update payment order status for order ID %d: %w", order.ID, err)
	}

	logger.GetLogger().Infof("Order ID %d on network %s updated to status %s with transferred amount: %s", order.ID, string(w.network), status, transferredAmount)
	w.processedOrderIDs[order.ID] = struct{}{}
	return nil
}

func (w *expiredOrderCatchupWorker) unpackTransferEvent(vLog types.Log) (dto.TransferEventDTO, error) {
	var transferEvent dto.TransferEventDTO

	// Ensure the number of topics matches the expected event (Transfer has 3 topics: event signature, from, to)
	if len(vLog.Topics) != 3 {
		return transferEvent, fmt.Errorf("invalid number of topics in log")
	}

	// The first topic is the event signature, so we skip it.
	// The second topic is the "from" address, and the third is the "to" address.
	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	// Unpack the value (the non-indexed parameter) from the data field
	err := w.parsedABI.UnpackIntoInterface(&transferEvent, constants.TransferEventName, vLog.Data)
	if err != nil {
		logger.GetLogger().Errorf("Failed to unpack transfer event on network %s: %v", string(w.network), err)
		return transferEvent, err
	}

	return transferEvent, nil
}
