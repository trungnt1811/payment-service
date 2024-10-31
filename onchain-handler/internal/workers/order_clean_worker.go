package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/log"
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
			log.LG.Info("Shutting down orderClearWorker")
			return
		}
	}
}

func (w *orderCleanWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		log.LG.Warn("Previous orderClearWorker run still in progress, skipping this cycle")
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
	err := w.paymentOrderUCase.UpdateExpiredOrdersToFailed(ctx)
	if err != nil {
		log.LG.Errorf("Failed to update expired orders to failed and release wallet: %v", err)
		return
	}
}

func (w *orderCleanWorker) updateActiveOrdersToExpired(ctx context.Context) {
	err := w.paymentOrderUCase.UpdateActiveOrdersToExpired(ctx)
	if err != nil {
		log.LG.Errorf("Failed to update active orders to expired: %v", err)
		return
	}
}
