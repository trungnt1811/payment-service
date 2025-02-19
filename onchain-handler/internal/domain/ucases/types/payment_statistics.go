package types

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type PaymentStatisticsUCase interface {
	IncrementStatistics(
		ctx context.Context,
		granularity string,
		periodStart time.Time,
		amount, transferred *string,
		symbol, vendorID string,
	) error
	GetStatisticsByTimeRangeAndGranularity(
		ctx context.Context,
		granularity string,
		startTime, endTime time.Time,
		vendorID string,
	) ([]dto.PaymentStatistics, error)
}
