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
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/log"
)

// expiredOrderCatchupWorker is a worker that processes expired orders.
type expiredOrderCatchupWorker struct {
	config                   *conf.Configuration
	paymentOrderUCase        interfaces.PaymentOrderUCase
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase
	cacheRepo                caching.CacheRepository
	contractAddress          string
	parsedABI                abi.ABI
	ethClient                *ethclient.Client
	isRunning                bool       // Tracks if catchup is running
	mu                       sync.Mutex // Mutex to protect the isRunning flag
}

func NewExpiredOrderCatchupWorker(
	config *conf.Configuration,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	cacheRepo caching.CacheRepository,
	contractAddress string,
	ethClient *ethclient.Client,
) interfaces.Worker {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		log.LG.Infof("failed to parse ERC20 ABI: %v", err)
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
			log.LG.Info("Shutting down expiredOrderCatchupWorker")
			return
		}
	}
}

// ensures that expiredOrderCatchupWorker doesn't overlap by checking the isRunning flag.
func (w *expiredOrderCatchupWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		log.LG.Warn("Previous catchupExpiredOrders run still in progress, skipping this cycle")
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
	expiredOrders, err := w.paymentOrderUCase.GetExpiredPaymentOrders(ctx)
	if err != nil {
		log.LG.Errorf("Failed to retrieve expired payment orders: %v", err)
		return
	}

	// No expired orders to process
	if len(expiredOrders) == 0 {
		log.LG.Info("No expired orders found")
		return
	}

	// Find the smallest block height from the expired orders
	minBlockHeight := expiredOrders[0].BlockHeight

	// Start processing logs from the smallest block height
	w.processExpiredOrders(ctx, minBlockHeight, expiredOrders)
}

// processExpiredOrders processes logs from the blockchain starting from the given block height
func (w *expiredOrderCatchupWorker) processExpiredOrders(ctx context.Context, startBlock uint64, expiredOrders []dto.PaymentOrderDTO) {
	log.LG.Infof("Processing expired orders starting from block: %d", startBlock)

	// Define the maximum block we should query up to
	latestBlock, err := w.getLatestBlockFromCacheOrBlockchain(ctx)
	if err != nil {
		log.LG.Errorf("Failed to retrieve latest block number: %v", err)
		return
	}

	// Calculate the effective latest block considering the confirmation depth.
	effectiveLatestBlock := latestBlock - constants.ConfirmationDepth

	// Check for invalid effective block height
	if effectiveLatestBlock <= 0 {
		log.LG.Warnf("Effective latest block (%d) is non-positive. Skipping processing for expired orders.", effectiveLatestBlock)
		return
	}

	// Safeguard if startBlock is beyond the effective latest block
	if startBlock > effectiveLatestBlock {
		log.LG.Warnf("Start block %d is beyond the effective latest block %d. No logs to process.", startBlock, effectiveLatestBlock)
		return
	}

	// Process logs in chunks of DefaultBlockOffset
	address := common.HexToAddress(w.contractAddress)
	for chunkStart := startBlock; chunkStart <= effectiveLatestBlock; chunkStart += constants.DefaultBlockOffset {
		chunkEnd := chunkStart + constants.DefaultBlockOffset - 1
		if chunkEnd > effectiveLatestBlock {
			chunkEnd = effectiveLatestBlock
		}

		log.LG.Debugf("Expired Order Catchup Worker: Processing block chunk from %d to %d", chunkStart, chunkEnd)

		// Poll logs from blockchain for this block range
		logs, err := utils.PollForLogsFromBlock(ctx, w.ethClient, []common.Address{address}, chunkStart, chunkEnd)
		if err != nil {
			log.LG.Errorf("Failed to poll logs from block range %d-%d: %v", chunkStart, chunkEnd, err)
			continue
		}

		// Process each log entry and match with expired orders
		for _, logEntry := range logs {
			err := w.processLog(ctx, logEntry, expiredOrders, logEntry.BlockNumber)
			if err != nil {
				log.LG.Errorf("Error processing log entry: %v", err)
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
	log.LG.Infof("Processing log entry from address: %s", vLog.Address.Hex())

	tokenSymbol, err := w.config.GetTokenSymbol(vLog.Address.Hex())
	if err != nil {
		return fmt.Errorf("failed to get token symbol from token contract address: %w", err)
	}

	// Unpack the transfer event from the log
	transferEvent, err := w.unpackTransferEvent(vLog)
	if err != nil {
		return fmt.Errorf("failed to unpack transfer event: %w", err)
	}

	// Iterate over all expired orders to find a matching wallet address
	for index, order := range orders {
		// Check if the event's "To" address matches the order's wallet address
		if order.Status != constants.Success && w.isMatchingWalletAddress(transferEvent.To.Hex(), order.Wallet.Address) && strings.EqualFold(order.Symbol, tokenSymbol) {
			// Found a matching order, now process the payment for that order
			log.LG.Infof("Matched transfer to wallet %s for order ID: %d", transferEvent.To.Hex(), order.ID)

			// Call processOrderPayment to handle the order update logic based on the transfer event
			isUpdated, err := w.processOrderPayment(ctx, &orders[index], transferEvent, blockHeight)
			if err != nil {
				return fmt.Errorf("failed to process order payment for order ID %d: %w", order.ID, err)
			}

			if isUpdated {
				if err := w.createPaymentEventHistory(ctx, order, transferEvent, tokenSymbol, vLog.Address.Hex(), vLog.TxHash.Hex()); err != nil {
					return fmt.Errorf("failed to create payment event history for order ID %d: %w", order.ID, err)
				}
				log.LG.Infof("Successfully processed order ID: %d with transferred amount: %s", order.ID, transferEvent.Value.String())
			}

			log.LG.Infof("Successfully processed order ID: %d with transferred amount: %s", order.ID, transferEvent.Value.String())
			return nil // Stop once we've processed the matching order
		}
	}

	// No matching order found for this log entry
	log.LG.Warnf("No matching expired order found for transfer to address: %s", transferEvent.To.Hex())
	return nil
}

// createPaymentEventHistory constructs and stores the payment event history
func (w *expiredOrderCatchupWorker) createPaymentEventHistory(
	ctx context.Context,
	order dto.PaymentOrderDTO,
	transferEvent dto.TransferEventDTO,
	tokenSymbol, contractAddress, txHash string,
) error {
	transferEventValueInEth, err := utils.ConvertWeiToEth(transferEvent.Value.String())
	if err != nil {
		log.LG.Errorf("Failed to convert transfer event value to ETH for order ID %d: %v", order.ID, err)
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
		log.LG.Infof("Processed order: %d. Ignore this turn.", order.ID)
		return false, nil
	}
	// Convert order amount and transferred amount into the appropriate unit (e.g., wei)
	orderAmount, err := utils.ConvertFloatEthToWei(order.Amount)
	if err != nil {
		return false, fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := utils.CalculatePaymentCovering(orderAmount, w.config.GetPaymentCovering())

	transferredAmount, err := utils.ConvertFloatEthToWei(order.Transferred)
	if err != nil {
		return false, fmt.Errorf("failed to convert transferred amount to wei: %v", err)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred := new(big.Int).Add(transferredAmount, transferEvent.Value)

	// Tracks whether the associated wallet is still in use.
	var inUse bool

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		log.LG.Infof("Processed full payment for order ID: %d", order.ID)

		// Set to false as the wallet is no longer in use after full payment (release wallet).
		inUse = false

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return true, w.updatePaymentOrderStatus(ctx, order, constants.Success, totalTransferred.String(), inUse, blockHeight)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		log.LG.Infof("Processed partial payment for order ID: %d", order.ID)

		// Set to true as the wallet is still in use for further payments.
		inUse = true

		// Update the order tranferred and keep the wallet associated with the order.
		return true, w.updatePaymentOrderStatus(ctx, order, order.Status, totalTransferred.String(), inUse, blockHeight)
	}

	return false, nil
}

// updatePaymentOrderStatus updates the payment order with the new status and transferred amount.
func (w *expiredOrderCatchupWorker) updatePaymentOrderStatus(
	ctx context.Context,
	order *dto.PaymentOrderDTO,
	status, transferredAmount string,
	inUse bool,
	blockHeight uint64,
) error {
	// Convert transferredAmount from Wei to Eth (Ether)
	transferredAmountInAvax, err := utils.ConvertWeiToEth(transferredAmount)
	if err != nil {
		return fmt.Errorf("updatePaymentOrderStatus error: %v", err)
	}

	// Update item in list
	order.Transferred = transferredAmountInAvax
	order.Status = status

	// Save the updated order to the repository
	err = w.paymentOrderUCase.UpdatePaymentOrder(ctx, order.ID, status, transferredAmountInAvax, inUse, blockHeight)
	if err != nil {
		return fmt.Errorf("failed to update payment order status for order ID %d: %w", order.ID, err)
	}

	log.LG.Infof("Order ID %d updated to status %s with transferred amount: %s", order.ID, status, transferredAmount)
	return nil
}

func (w *expiredOrderCatchupWorker) isMatchingWalletAddress(eventToAddress, orderWallet string) bool {
	return strings.EqualFold(eventToAddress, orderWallet)
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
		log.LG.Errorf("Failed to unpack transfer event: %v", err)
		return transferEvent, err
	}

	return transferEvent, nil
}

func (w *expiredOrderCatchupWorker) getLatestBlockFromCacheOrBlockchain(ctx context.Context) (uint64, error) {
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey}

	var latestBlock uint64
	err := w.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err == nil {
		log.LG.Debugf("Retrieved latest block number from cache: %d", latestBlock)
		return latestBlock, nil
	}

	// If cache is empty, load from blockchain
	latest, err := utils.GetLatestBlockNumber(ctx, w.ethClient)
	if err != nil {
		log.LG.Errorf("Failed to retrieve the latest block number from blockchain: %v", err)
		return 0, err
	}

	log.LG.Debugf("Retrieved latest block number from blockchain: %d", latest.Uint64())
	return latest.Uint64(), nil
}
