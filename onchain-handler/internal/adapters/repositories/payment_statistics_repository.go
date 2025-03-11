package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type paymentStatisticsRepository struct {
	db *gorm.DB
}

func NewPaymentStatisticsRepository(db *gorm.DB) repotypes.PaymentStatisticsRepository {
	return &paymentStatisticsRepository{
		db: db,
	}
}

// IncrementStatistics increments or initializes statistics for a specific granularity, period, symbol, and vendor.
func (r *paymentStatisticsRepository) IncrementStatistics(
	ctx context.Context,
	granularity string,
	periodStart time.Time,
	amount, transferred *string,
	symbol, vendorID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if amount != nil {
			updates["total_orders"] = gorm.Expr("total_orders + 1")
			updates["total_amount"] = gorm.Expr("total_amount::numeric + ?", *amount)
		}
		if transferred != nil {
			updates["total_transferred"] = gorm.Expr("total_transferred::numeric + ?", *transferred)
		}

		// Attempt to update with row-level locking
		result := tx.Model(&entities.PaymentStatistics{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("granularity = ? AND period_start = ? AND vendor_id = ? AND symbol = ?", granularity, periodStart.UTC(), vendorID, symbol).
			Updates(updates)

		if result.Error != nil {
			return fmt.Errorf("failed to update payment statistics: %w", result.Error)
		}

		// If no rows were updated, insert a new record
		if result.RowsAffected == 0 {
			newStatistic := entities.PaymentStatistics{
				Granularity:      granularity,
				PeriodStart:      periodStart.UTC(),
				TotalOrders:      0,
				TotalAmount:      "0",
				TotalTransferred: "0",
				Symbol:           symbol,
				VendorID:         vendorID,
			}

			if amount != nil {
				newStatistic.TotalOrders = 1
				newStatistic.TotalAmount = *amount
			}
			if transferred != nil {
				newStatistic.TotalTransferred = *transferred
			}

			if err := tx.Create(&newStatistic).Error; err != nil {
				return fmt.Errorf("failed to insert new payment statistics: %w", err)
			}
		}
		return nil
	})
}

func (r *paymentStatisticsRepository) GetStatisticsByTimeRangeAndGranularity(
	ctx context.Context,
	granularity string,
	startTime, endTime time.Time,
	vendorID string,
) ([]entities.PaymentStatistics, error) {
	var statistics []entities.PaymentStatistics

	// Handle grouping for granularities other than DAILY
	query := r.db.WithContext(ctx).Model(&entities.PaymentStatistics{}).Order("period_start ASC")

	// Modify query based on granularity
	switch granularity {
	case constants.Daily:
		// No grouping needed for DAILY
		query = query.Where("granularity = ? AND period_start >= ? AND period_start < ? AND vendor_id = ?", granularity, startTime, endTime, vendorID)

	default:
		// Apply grouping for WEEKLY, MONTHLY, and YEARLY
		groupByUnit := ""
		switch granularity {
		case constants.Weekly:
			groupByUnit = "week"
		case constants.Monthly:
			groupByUnit = "month"
		case constants.Yearly:
			groupByUnit = "year"
		default:
			return nil, fmt.Errorf("unsupported granularity: %s", granularity)
		}

		query = query.Select([]string{
			fmt.Sprintf("DATE_TRUNC('%s', period_start) AS period_start", groupByUnit),
			"SUM(total_orders) AS total_orders",
			"SUM(total_amount::numeric) AS total_amount",
			"SUM(total_transferred::numeric) AS total_transferred",
			"symbol",
			"vendor_id",
		}).
			Where("granularity = ? AND period_start >= ? AND period_start < ? AND vendor_id = ?", constants.Daily, startTime, endTime, vendorID).
			Group("symbol, vendor_id, DATE_TRUNC('" + groupByUnit + "', period_start)")
	}

	// Execute query
	if err := query.Find(&statistics).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment statistics: %w", err)
	}

	return statistics, nil
}
