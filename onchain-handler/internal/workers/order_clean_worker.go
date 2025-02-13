package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
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

	errors := utils.SendWebhooks(
		ctx,
		utils.ToInterfaceSlice(orderDTOs),
		func(order interface{}) string {
			return order.(dto.PaymentOrderDTOResponse).WebhookURL
		},
	)
	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	} else {
		logger.GetLogger().Info("All webhooks for failed orders sent successfully.")
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

	errors := utils.SendWebhooks(
		ctx,
		utils.ToInterfaceSlice(orderDTOs),
		func(order interface{}) string {
			return order.(dto.PaymentOrderDTOResponse).WebhookURL
		},
	)
	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for some orders: %v", errors)
	} else {
		logger.GetLogger().Info("All webhooks for expired orders sent successfully.")
	}
}
