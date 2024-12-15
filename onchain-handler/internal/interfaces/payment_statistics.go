package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentStatisticsRepository interface {
	IncrementStatistics(
		ctx context.Context,
		granularity string,
		periodStart time.Time,
		amount, transferred *string,
		symbol, vendorID string,
	) error
	AggregateToHigherGranularities(ctx context.Context) error
	GetStatisticsByTimeRangeAndGranularity(
		ctx context.Context,
		granularity string,
		startTime, endTime time.Time,
		vendorID string,
	) ([]domain.PaymentStatistics, error)
}

type PaymentStatisticsUCase interface {
	IncrementStatistics(
		ctx context.Context,
		granularity string,
		periodStart time.Time,
		amount, transferred *string,
		symbol, vendorID string,
	) error
	AggregateToHigherGranularities(ctx context.Context) error
	GetStatisticsByTimeRangeAndGranularity(
		ctx context.Context,
		granularity string,
		startTime, endTime time.Time,
		vendorID string,
	) ([]dto.PaymentStatistics, error)
}
