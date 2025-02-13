package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
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
	GetStatisticsByTimeRangeAndGranularity(
		ctx context.Context,
		granularity string,
		startTime, endTime time.Time,
		vendorID string,
	) ([]entities.PaymentStatistics, error)
}

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
