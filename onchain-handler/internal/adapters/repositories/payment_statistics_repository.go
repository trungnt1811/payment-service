package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentStatisticsRepository struct {
	db *gorm.DB
}

func NewPaymentStatisticsRepository(db *gorm.DB) interfaces.PaymentStatisticsRepository {
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
		updates := map[string]interface{}{}
		if amount != nil {
			updates["total_orders"] = gorm.Expr("total_orders + 1")
			updates["total_amount"] = gorm.Expr("total_amount::numeric + ?", *amount)
		}
		if transferred != nil {
			updates["total_transferred"] = gorm.Expr("total_transferred::numeric + ?", *transferred)
		}

		// Attempt to update with row-level locking
		result := tx.Model(&domain.PaymentStatistics{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("granularity = ? AND period_start = ? AND vendor_id = ? AND symbol = ?", granularity, periodStart.UTC(), vendorID, symbol).
			Updates(updates)

		if result.Error != nil {
			return fmt.Errorf("failed to update payment statistics: %w", result.Error)
		}

		// If no rows were updated, insert a new record
		if result.RowsAffected == 0 {
			newStatistic := domain.PaymentStatistics{
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

func (r *paymentStatisticsRepository) AggregateToHigherGranularities(ctx context.Context) error {
	// Define the aggregation chain
	aggregationChains := []struct {
		Source constants.Granularity
		Target constants.Granularity
	}{
		{constants.Daily, constants.Weekly},
		{constants.Weekly, constants.Monthly},
		{constants.Monthly, constants.Yearly},
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Loop through each aggregation chain and process them
		for _, chain := range aggregationChains {
			if err := r.aggregateGranularity(tx, chain.Source, chain.Target); err != nil {
				return fmt.Errorf("failed to aggregate %s to %s: %w", chain.Source, chain.Target, err)
			}
		}
		return nil
	})
}

func (r *paymentStatisticsRepository) aggregateGranularity(
	tx *gorm.DB,
	sourceGranularity, targetGranularity constants.Granularity,
) error {
	// Define a temporary struct to hold the aggregated results
	type AggregatedResult struct {
		TotalOrders      uint64
		TotalAmount      string
		TotalTransferred string
		Symbol           string
		VendorID         string
		PeriodStart      time.Time
	}

	// Determine the cutoff point to exclude incomplete periods
	cutoffTime := utils.GetPeriodStart(string(targetGranularity), time.Now())

	// Lock the source rows for aggregation to prevent concurrent processing
	var sourceRecords []domain.PaymentStatistics
	if err := tx.Model(&domain.PaymentStatistics{}).
		Clauses(clause.Locking{Strength: "UPDATE"}). // Apply row-level locking
		Where("granularity = ? AND is_aggregated = ? AND period_start < ?", sourceGranularity, false, cutoffTime).
		Find(&sourceRecords).Error; err != nil {
		return fmt.Errorf("failed to lock source records for granularity %s: %w", sourceGranularity, err)
	}

	// Perform aggregation on locked records
	var results []AggregatedResult
	err := tx.Model(&domain.PaymentStatistics{}).
		Select([]string{
			"SUM(total_orders) AS total_orders",
			"SUM(total_amount::numeric) AS total_amount",
			"SUM(total_transferred::numeric) AS total_transferred",
			"symbol",
			"vendor_id",
			"MIN(period_start) AS period_start",
			fmt.Sprintf("DATE_TRUNC('%s', period_start) AS truncated_period", targetGranularity.ToDateTruncUnit()),
		}).
		Where("granularity = ? AND is_aggregated = ? AND period_start < ?", sourceGranularity, false, cutoffTime).
		Group("symbol, vendor_id, truncated_period").
		Find(&results).Error
	if err != nil {
		return fmt.Errorf("failed to aggregate data: %w", err)
	}

	// Upsert the aggregated results into the target granularity
	for _, result := range results {
		periodStart := utils.GetPeriodStart(string(targetGranularity), result.PeriodStart)

		// Try updating the existing record
		updateResult := tx.Model(&domain.PaymentStatistics{}).
			Where("granularity = ? AND period_start = ? AND symbol = ? AND vendor_id = ?",
				targetGranularity, periodStart.UTC(), result.Symbol, result.VendorID).
			Updates(map[string]interface{}{
				"total_orders":      gorm.Expr("total_orders + ?", result.TotalOrders),
				"total_amount":      gorm.Expr("total_amount::numeric + ?", result.TotalAmount),
				"total_transferred": gorm.Expr("total_transferred::numeric + ?", result.TotalTransferred),
			})

		if updateResult.Error != nil {
			return fmt.Errorf("failed to update payment statistics: %w", updateResult.Error)
		}

		// Insert a new record if no row was updated
		if updateResult.RowsAffected == 0 {
			newStat := domain.PaymentStatistics{
				Granularity:      string(targetGranularity),
				PeriodStart:      periodStart.UTC(),
				TotalOrders:      result.TotalOrders,
				TotalAmount:      result.TotalAmount,
				TotalTransferred: result.TotalTransferred,
				Symbol:           result.Symbol,
				VendorID:         result.VendorID,
			}

			if err := tx.Create(&newStat).Error; err != nil {
				return fmt.Errorf("failed to insert new aggregated data: %w", err)
			}
		}
	}

	// Mark the source records as aggregated
	err = tx.Model(&domain.PaymentStatistics{}).
		Where("granularity = ? AND is_aggregated = ? AND period_start < ?", sourceGranularity, false, cutoffTime).
		Update("is_aggregated", true).Error
	if err != nil {
		return fmt.Errorf("failed to update is_aggregated flag for %s: %w", sourceGranularity, err)
	}

	return nil
}

func (r *paymentStatisticsRepository) GetStatisticsByTimeRangeAndGranularity(
	ctx context.Context,
	granularity string,
	startTime, endTime time.Time,
	vendorID string,
) ([]domain.PaymentStatistics, error) {
	var statistics []domain.PaymentStatistics

	// Perform the query with the given parameters
	err := r.db.WithContext(ctx).
		Model(&domain.PaymentStatistics{}).
		Where("granularity = ? AND period_start >= ? AND period_start < ? AND vendor_id = ?", granularity, startTime, endTime, vendorID).
		Order("period_start ASC"). // Order by period_start for better visualization
		Find(&statistics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment statistics: %w", err)
	}

	return statistics, nil
}
