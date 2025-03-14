package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	workertypes "github.com/genefriendway/onchain-handler/internal/workers/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type orderCleanWorker struct {
	paymentOrderUCase ucasetypes.PaymentOrderUCase
	isRunning         bool
	mu                sync.Mutex
}

func NewOrderCleanWorker(paymentOrderUCase ucasetypes.PaymentOrderUCase) workertypes.Worker {
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
	w.updateOrdersAndSendWebhooks(
		ctx,
		w.paymentOrderUCase.UpdateExpiredOrdersToFailed,
		"update expired orders to failed and release wallet",
		"All webhooks for failed orders sent successfully.",
		"No expired orders updated to failed.",
	)
}

func (w *orderCleanWorker) updateActiveOrdersToExpired(ctx context.Context) {
	w.updateOrdersAndSendWebhooks(
		ctx,
		w.paymentOrderUCase.UpdateActiveOrdersToExpired,
		"update active orders to expired",
		"All webhooks for expired orders sent successfully.",
		"No active orders updated to expired.",
	)
}

func (w *orderCleanWorker) updateOrdersAndSendWebhooks(
	ctx context.Context,
	updateFunc func(context.Context) ([]uint64, error),
	actionDesc string,
	successLog string,
	noActionDesc string,
) {
	orderIDs, err := updateFunc(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to %s: %v", actionDesc, err)
		return
	}

	if len(orderIDs) == 0 {
		logger.GetLogger().Info(noActionDesc)
		return
	}

	orderDTOs, err := w.paymentOrderUCase.GetPaymentOrdersByIDs(ctx, orderIDs)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment orders for IDs %v: %v", orderIDs, err)
		return
	}

	errors := utils.SendWebhooks(
		ctx,
		utils.ToInterfaceSlice(orderDTOs),
		func(order any) string {
			return order.(dto.PaymentOrderDTOResponse).WebhookURL
		},
	)

	if len(errors) > 0 {
		logger.GetLogger().Errorf("Failed to send webhooks for orders %v: %v", orderIDs, errors)
	} else {
		logger.GetLogger().Info(successLog)
	}
}
