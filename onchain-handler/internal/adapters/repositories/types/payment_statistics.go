package types

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type PaymentStatisticsRepository interface {
	IncrementStatistics(
		ctx context.Context,
		granularity string,
		periodStart time.Time,
		amount, transferred *string,
		symbol, vendorID string,
	) error
	RevertAndIncrementStatistics(
		ctx context.Context,
		granularity string,
		periodStart time.Time,
		amount *string,
		oldSymbol, newSymbol, vendorID string,
	) error
	GetStatisticsByTimeRangeAndGranularity(
		ctx context.Context,
		granularity string,
		startTime, endTime time.Time,
		vendorID string,
		symbols []string,
	) ([]entities.PaymentStatistics, error)
}
