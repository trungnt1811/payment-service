package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type orderCleanWorker struct {
	paymentOrderUCase interfaces.PaymentOrderUCase
	isRunning         bool
	mu                sync.Mutex
}

func NewOrderCleanWorker(paymentOrderUCase interfaces.PaymentOrderUCase) interfaces.Worker {
	return &orderCleanWorker{
		paymentOrderUCase: paymentOrderUCase,
	}
}

func (w *orderCleanWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.OrderCleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx)
		case <-ctx.Done():
			logger.GetLogger().Info("Shutting down orderCleanWorker")
			return
		}
	}
}

func (w *orderCleanWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warn("Previous orderCleanWorker run still in progress, skipping this cycle")
		w.mu.Unlock()
		return
	}

	// Mark as running
	w.isRunning = true
	w.mu.Unlock()

	w.releaseWallet(ctx)
	w.updateActiveOrdersToExpired(ctx)

	// Mark as not running
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()
}

func (w *orderCleanWorker) releaseWallet(ctx context.Context) {
	orderIDs, err := w.paymentOrderUCase.UpdateExpiredOrdersToFailed(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to update expired orders to failed and release wallet: %v", err)
		return
	}

	orderDTOs, err := w.paymentOrderUCase.GetPaymentOrdersByIDs(ctx, orderIDs)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get payment orders by IDs: %v", err)
		return
	}

	// Limit concurrency with a semaphore
	sem := make(chan struct{}, constants.MaxWebhookWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, orderDTO := range orderDTOs {
		sem <- struct{}{} // Acquire a slot
		wg.Add(1)

		go func(order dto.PaymentOrderDTOResponse) {
			defer func() {
				<-sem // Release the slot
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				logger.GetLogger().Warnf("Context canceled before sending webhook for order ID %d", order.ID)
				return
			default:
				if err := utils.SendWebhook(order, order.WebhookURL); err != nil {
					logger.GetLogger().Errorf("Failed to send webhook for order ID %d: %v", order.ID, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("order ID %d: %w", order.ID, err))
					mu.Unlock()
				} else {
					logger.GetLogger().Infof("Webhook sent successfully for order ID %d", order.ID)
				}
			}
		}(orderDTO)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	}
}

func (w *orderCleanWorker) updateActiveOrdersToExpired(ctx context.Context) {
	orderIDs, err := w.paymentOrderUCase.UpdateActiveOrdersToExpired(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to update active orders to expired: %v", err)
		return
	}

	if len(orderIDs) == 0 {
		logger.GetLogger().Info("No active orders updated to expired.")
		return
	}

	orderDTOs, err := w.paymentOrderUCase.GetPaymentOrdersByIDs(ctx, orderIDs)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get payment orders by IDs: %v", err)
		return
	}

	// Limit concurrency with a semaphore
	sem := make(chan struct{}, constants.MaxWebhookWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, orderDTO := range orderDTOs {
		sem <- struct{}{} // Acquire a slot
		wg.Add(1)

		go func(order dto.PaymentOrderDTOResponse) {
			defer func() {
				<-sem // Release the slot
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				logger.GetLogger().Warnf("Context canceled before sending webhook for order ID %d", order.ID)
				return
			default:
				if err := utils.SendWebhook(order, order.WebhookURL); err != nil {
					logger.GetLogger().Errorf("Failed to send webhook for order ID %d: %v", order.ID, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("order ID %d: %w", order.ID, err))
					mu.Unlock()
				} else {
					logger.GetLogger().Infof("Webhook sent successfully for order ID %d", order.ID)
				}
			}
		}(orderDTO)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	} else {
		logger.GetLogger().Info("All webhooks for updated orders sent successfully.")
	}
}
