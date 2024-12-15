package workers

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type paymentStatisticsWorker struct {
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase
}

func NewPaymentStatisticsWorker(
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
) interfaces.Worker {
	return &paymentStatisticsWorker{
		paymentStatisticsUCase: paymentStatisticsUCase,
	}
}

func (w *paymentStatisticsWorker) Start(ctx context.Context) {
	for {
		// Calculate the duration until the next scheduled time (e.g., midnight)
		now := time.Now()
		nextRun := time.Date(
			now.Year(), now.Month(), now.Day()+1, // Next day
			0, 0, 0, 0, // 00:00:00
			time.UTC, // Use UTC for consistent timing
		)
		sleepDuration := time.Until(nextRun)
		logger.GetLogger().Infof("Next payment statistics aggregation scheduled at: %s", nextRun)
		// Sleep until the next scheduled time or exit early if the context is canceled
		select {
		case <-time.After(sleepDuration):
			w.run(ctx) // Execute the aggregation task
		case <-ctx.Done():
			logger.GetLogger().Info("Shutting down paymentStatisticsWorker")
			return
		}
	}
}

func (w *paymentStatisticsWorker) run(ctx context.Context) {
	logger.GetLogger().Info("Starting payment statistics aggregation")
	err := w.paymentStatisticsUCase.AggregateToHigherGranularities(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to aggregate payment statistics: %v", err)
	} else {
		logger.GetLogger().Info("Successfully completed payment statistics aggregation")
	}
}
