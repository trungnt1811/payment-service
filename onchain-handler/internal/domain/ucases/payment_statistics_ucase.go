package ucases

import (
	"context"
	"time"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

type paymentStatisticsUCase struct {
	paymentStatisticsRepository repotypes.PaymentStatisticsRepository
}

func NewPaymentStatisticsCase(
	paymentStatisticsRepository repotypes.PaymentStatisticsRepository,
) ucasetypes.PaymentStatisticsUCase {
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

// GetStatisticsByTimeRangeAndGranularity retrieves payment statistics by time range and granularity.
func (u *paymentStatisticsUCase) GetStatisticsByTimeRangeAndGranularity(
	ctx context.Context,
	granularity string,
	startTime, endTime time.Time,
	vendorID string,
	symbols []string,
) ([]dto.PeriodStatistics, error) {
	// Retrieve payment statistics from the repository
	paymentStatistics, err := u.paymentStatisticsRepository.GetStatisticsByTimeRangeAndGranularity(
		ctx, granularity, startTime, endTime, vendorID, symbols,
	)
	if err != nil {
		return nil, err
	}

	// Convert the payment statistics to DTO format
	return entities.ToPeriodStatisticsDTO(paymentStatistics), nil
}
