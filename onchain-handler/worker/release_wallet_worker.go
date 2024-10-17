package worker

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

type releaseWalletWorker struct {
	paymentOrderUCase interfaces.PaymentOrderUCase
	isRunning         bool
	mu                sync.Mutex
}

func NewReleaseWalletWorker(paymentOrderUCase interfaces.PaymentOrderUCase) Worker {
	return &releaseWalletWorker{
		paymentOrderUCase: paymentOrderUCase,
	}
}

func (w *releaseWalletWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.ReleaseWalletInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx)
		case <-ctx.Done():
			log.LG.Info("Shutting down releaseWalletWorker")
			return
		}
	}
}

func (w *releaseWalletWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		log.LG.Warn("Previous releaseWalletWorker run still in progress, skipping this cycle")
		w.mu.Unlock()
		return
	}

	// Mark as running
	w.isRunning = true
	w.mu.Unlock()

	// Perform the release wallet process
	w.releaseWallet(ctx)

	// Mark as not running
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()
}

func (w *releaseWalletWorker) releaseWallet(ctx context.Context) {
	err := w.paymentOrderUCase.UpdateExpiredOrdersToFailed(ctx)
	if err != nil {
		log.LG.Errorf("Failed to update expired orders to failed and release wallet: %v", err)
		return
	}
}
