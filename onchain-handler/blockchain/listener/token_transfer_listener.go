package listener

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

	"github.com/genefriendway/onchain-handler/blockchain/event"
	listenerinterfaces "github.com/genefriendway/onchain-handler/blockchain/interfaces"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
	"github.com/genefriendway/onchain-handler/utils"
)

// tokenTransferListener listens for token transfers and processes them using a queue of payment orders.
type tokenTransferListener struct {
	ctx               context.Context
	config            *conf.Configuration
	baseEventListener listenerinterfaces.BaseEventListener
	paymentOrderUCase interfaces.PaymentOrderUCase
	contractAddress   string
	parsedABI         abi.ABI
	queue             *queue.Queue[dto.PaymentOrderDTO]
	mu                sync.Mutex // Mutex for ticker synchronization
}

// NewTokenTransferListener creates a new tokenTransferListener with a payment order queue.
func NewTokenTransferListener(
	ctx context.Context,
	config *conf.Configuration,
	baseEventListener listenerinterfaces.BaseEventListener,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	contractAddress string,
	orderQueue *queue.Queue[dto.PaymentOrderDTO],
) (listenerinterfaces.EventListener, error) {
	parsedABI, err := abi.JSON(strings.NewReader(constants.Erc20TransferEventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	listener := &tokenTransferListener{
		ctx:               ctx,
		config:            config,
		baseEventListener: baseEventListener,
		paymentOrderUCase: paymentOrderUCase,
		contractAddress:   contractAddress,
		queue:             orderQueue,
		parsedABI:         parsedABI,
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
			log.LG.Info("Starting dequeue operation...")
			listener.dequeueOrders()
			log.LG.Info("Dequeue operation finished.")
			listener.mu.Unlock()
		case <-listener.ctx.Done():
			log.LG.Infof("Stopping dequeue ticker: %v", listener.ctx.Err())
			return
		}
	}
}

// parseAndProcessTransferEvent parses and processes a transfer event, checking if it matches any payment order in the queue.
func (listener *tokenTransferListener) parseAndProcessTransferEvent(vLog types.Log) (interface{}, error) {
	// Handle expired orders and success orders before processing the event
	listener.dequeueOrders()

	// Process the event
	transferEvent, err := listener.unpackTransferEvent(vLog)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack transfer event: %w", err)
	}

	log.LG.Infof("Detected Transfer event: From: %s, To: %s, Value: %s", transferEvent.From.Hex(), transferEvent.To.Hex(), transferEvent.Value.String())

	queueItems := listener.queue.GetItems()

	// Process payment orders based on the transfer event
	for index, order := range queueItems {
		// Check if transfer matches the order's wallet and token symbol
		if listener.isMatchingWalletAddress(transferEvent.To.Hex(), order.PaymentAddress) &&
			listener.isMatchingTokenSymbol(transferEvent.From.Hex(), order.Symbol) {
			if err := listener.processOrderPayment(&queueItems[index], transferEvent, vLog.BlockNumber); err != nil {
				log.LG.Errorf("Failed to process payment for order ID: %d, error: %v", order.ID, err)
			}
		}
	}

	return transferEvent, nil
}

func (listener *tokenTransferListener) unpackTransferEvent(vLog types.Log) (event.TransferEvent, error) {
	var transferEvent event.TransferEvent

	// Ensure the number of topics matches the expected event (Transfer has 3 topics: event signature, from, to)
	if len(vLog.Topics) != 3 {
		return transferEvent, fmt.Errorf("invalid number of topics in log")
	}

	// The first topic is the event signature, so we skip it.
	// The second topic is the "from" address, and the third is the "to" address.
	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	// Unpack the value (the non-indexed parameter) from the data field
	err := listener.parsedABI.UnpackIntoInterface(&transferEvent, constants.TransferEventName, vLog.Data)
	if err != nil {
		log.LG.Errorf("Failed to unpack transfer event: %v", err)
		return transferEvent, err
	}

	return transferEvent, nil
}

func (listener *tokenTransferListener) isMatchingTokenSymbol(tokenAddress, orderSymbol string) bool {
	// Get the token symbol for the provided address
	if tokenSymbol, err := listener.config.GetTokenSymbol(tokenAddress); err == nil && tokenSymbol == orderSymbol {
		return true
	}
	return false
}

func (listener *tokenTransferListener) isMatchingWalletAddress(eventToAddress, orderWalletAddress string) bool {
	return strings.EqualFold(eventToAddress, orderWalletAddress)
}

// processOrderPayment handles the payment for an order based on the transfer event details.
// It updates the order status and wallet usage based on the payment amount.
func (listener *tokenTransferListener) processOrderPayment(order *dto.PaymentOrderDTO, transferEvent event.TransferEvent, blockHeight uint64) error {
	orderAmount, err := utils.ConvertFloatEthToWei(order.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert order amount: %v", err)
	}
	minimumAcceptedAmount := utils.CalculatePaymentCovering(orderAmount, listener.config.GetPaymentCovering())

	transferredAmount, err := utils.ConvertFloatEthToWei(order.Transferred)
	if err != nil {
		return fmt.Errorf("failed to convert transferred amount: %v", err)
	}

	// Calculate the total transferred amount by adding the new transfer event value.
	totalTransferred := new(big.Int).Add(transferredAmount, transferEvent.Value)

	// Tracks whether the associated wallet is still in use.
	var inUse bool

	// Check if the total transferred amount is greater than or equal to the minimum accepted amount (full payment).
	if totalTransferred.Cmp(minimumAcceptedAmount) >= 0 {
		log.LG.Infof("Processed full payment for order ID: %d", order.ID)

		// Set to false as the wallet is no longer in use after full payment.
		inUse = false

		// Update the order status to 'Success' and mark the wallet as no longer in use.
		return listener.updatePaymentOrderStatus(order, constants.Success, totalTransferred.String(), inUse, blockHeight)
	} else if totalTransferred.Cmp(big.NewInt(0)) > 0 {
		// If the total transferred amount is greater than 0 but less than the minimum accepted amount (partial payment).
		log.LG.Infof("Processed partial payment for order ID: %d", order.ID)

		// Set to true as the wallet is still in use for further payments.
		inUse = true

		// Update the order status to 'Partial' and keep the wallet associated with the order.
		return listener.updatePaymentOrderStatus(order, constants.Partial, totalTransferred.String(), inUse, blockHeight)
	}

	return nil
}

func (listener *tokenTransferListener) updatePaymentOrderStatus(
	order *dto.PaymentOrderDTO,
	status, transferredAmount string,
	walletStatus bool,
	blockHeight uint64,
) error {
	// Convert transferredAmount from Wei to Eth
	transferredAmountInEth, err := utils.ConvertWeiToEth(transferredAmount)
	if err != nil {
		return fmt.Errorf("updatePaymentOrderStatus error: %v", err)
	}

	// Also update item in queue
	order.Transferred = transferredAmountInEth
	order.Status = status

	// Call the method to update the payment order
	return listener.paymentOrderUCase.UpdatePaymentOrder(listener.ctx, order.ID, status, transferredAmountInEth, walletStatus, blockHeight)
}

// dequeueOrders removes expired or successful orders from the queue and refills it to maintain the limit.
func (listener *tokenTransferListener) dequeueOrders() {
	// Retrieve all current orders in the queue
	orders := listener.queue.GetItems()

	var dequeueOrders []dto.PaymentOrderDTO
	// Iterate over current orders in the queue
	for index, order := range orders {
		// Check if the order needs to be dequeued
		if listener.shouldDequeueOrder(order) {
			if order.Status != constants.Success {
				orders[index].Status = constants.Expired
				dequeueOrders = append(dequeueOrders, orders[index])
			}
			if err := listener.queue.Dequeue(func(o dto.PaymentOrderDTO) bool {
				return o.ID == order.ID
			}); err != nil {
				log.LG.Errorf("Failed to dequeue order ID: %d, error: %v", order.ID, err)
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
			log.LG.Errorf("Failed to update orders with IDs: %v, error: %v", failedOrderIDs, err)
		} else {
			// Log all successful order IDs
			var successfulOrderIDs []uint64
			for _, order := range dequeueOrders {
				successfulOrderIDs = append(successfulOrderIDs, order.ID)
			}
			log.LG.Infof("Successfully updated orders with IDs: %v to expired status", successfulOrderIDs)
		}
	}

	// Refill the queue to ensure it has the required number of items
	if err := listener.queue.FillQueue(); err != nil {
		log.LG.Errorf("Failed to refill the queue: %v", err)
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

// Register registers the listener for token transfer events.
func (listener *tokenTransferListener) Register(ctx context.Context) {
	listener.baseEventListener.RegisterEventListener(
		listener.contractAddress,
		listener.parseAndProcessTransferEvent,
	)
}
