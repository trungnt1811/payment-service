package ucases

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentStatisticsUCase struct {
	paymentStatisticsRepository interfaces.PaymentStatisticsRepository
}

func NewPaymentStatisticsCase(
	paymentStatisticsRepository interfaces.PaymentStatisticsRepository,
) interfaces.PaymentStatisticsUCase {
	return &paymentStatisticsUCase{
		paymentStatisticsRepository: paymentStatisticsRepository,
	}
}

func (u *paymentStatisticsUCase) IncrementStatistics(
	ctx context.Context,
	granularity string,
	periodStart time.Time,
	amount, transferred *string,
	symbol, vendorID string,
) error {
	return u.paymentStatisticsRepository.IncrementStatistics(
		ctx, granularity, periodStart, amount, transferred, symbol, vendorID,
	)
}

func (u *paymentStatisticsUCase) AggregateToHigherGranularities(ctx context.Context) error {
	return u.paymentStatisticsRepository.AggregateToHigherGranularities(ctx)
}

// GetStatisticsByTimeRangeAndGranularity retrieves payment statistics by time range and granularity.
func (u *paymentStatisticsUCase) GetStatisticsByTimeRangeAndGranularity(
	ctx context.Context,
	granularity string,
	startTime, endTime time.Time,
	vendorID string,
) ([]dto.PaymentStatistics, error) {
	// Retrieve payment statistics from the repository
	paymentStatistics, err := u.paymentStatisticsRepository.GetStatisticsByTimeRangeAndGranularity(
		ctx, granularity, startTime, endTime, vendorID,
	)
	if err != nil {
		return nil, err
	}

	// Convert domain models to DTOs
	var dtos []dto.PaymentStatistics
	for _, stat := range paymentStatistics {
		dtos = append(dtos, stat.ToDto())
	}

	return dtos, nil
}
