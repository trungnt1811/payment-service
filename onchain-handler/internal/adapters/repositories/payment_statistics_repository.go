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

var paymentStatsWhereClause = "granularity = ? AND period_start = ? AND vendor_id = ? AND symbol = ?"

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
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Where(paymentStatsWhereClause, granularity, periodStart.UTC(), vendorID, symbol).
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

func (r *paymentStatisticsRepository) RevertAndIncrementStatistics(
	ctx context.Context,
	granularity string,
	periodStart time.Time,
	amount *string,
	oldSymbol, newSymbol, vendorID string,
) error {
	if oldSymbol == newSymbol {
		return nil // no-op
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Revert old
		if _, err := r.updateStatsInTx(tx, granularity, periodStart, vendorID, oldSymbol, amount, -1); err != nil {
			return fmt.Errorf("failed to revert old payment statistics: %w", err)
		}

		// Step 2: Apply new stats
		rowsAffected, err := r.updateStatsInTx(tx, granularity, periodStart, vendorID, newSymbol, amount, 1)
		if err != nil {
			return fmt.Errorf("failed to increment new payment statistics: %w", err)
		}

		// Step 3: Insert if new stats didn't exist
		if rowsAffected == 0 {
			stat := entities.PaymentStatistics{
				Granularity:      granularity,
				PeriodStart:      periodStart.UTC(),
				TotalOrders:      0,
				TotalAmount:      "0",
				TotalTransferred: "0",
				Symbol:           newSymbol,
				VendorID:         vendorID,
			}
			if amount != nil {
				stat.TotalOrders = 1
				stat.TotalAmount = *amount
			}
			if err := tx.Create(&stat).Error; err != nil {
				return fmt.Errorf("failed to insert new payment statistics: %w", err)
			}
		}

		return nil
	})
}

// Helper to apply +/- update to stats
func (r *paymentStatisticsRepository) updateStatsInTx(
	tx *gorm.DB,
	granularity string,
	periodStart time.Time,
	vendorID, symbol string,
	amount *string,
	sign int,
) (int64, error) {
	updates := map[string]any{}
	if amount != nil {
		op := "+"
		if sign < 0 {
			op = "-"
		}
		updates["total_orders"] = gorm.Expr(fmt.Sprintf("total_orders %s 1", op))
		updates["total_amount"] = gorm.Expr(fmt.Sprintf("total_amount::numeric %s ?", op), *amount)
	}

	result := tx.Model(&entities.PaymentStatistics{}).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where(paymentStatsWhereClause, granularity, periodStart.UTC(), vendorID, symbol).
		Updates(updates)

	return result.RowsAffected, result.Error
}

func (r *paymentStatisticsRepository) GetStatisticsByTimeRangeAndGranularity(
	ctx context.Context,
	granularity string,
	startTime, endTime time.Time,
	vendorID string,
	symbols []string,
) ([]entities.PaymentStatistics, error) {
	var statistics []entities.PaymentStatistics

	query := r.db.WithContext(ctx).Model(&entities.PaymentStatistics{}).Order("period_start ASC")

	// Common WHERE condition
	baseConditions := query.Where(
		"period_start >= ? AND period_start < ? AND vendor_id = ?",
		startTime.UTC(), endTime.UTC(), vendorID,
	)

	// Optional symbol filtering
	if len(symbols) > 0 {
		baseConditions = baseConditions.Where("symbol IN ?", symbols)
	}

	switch granularity {
	case constants.Daily:
		// Use raw daily stats
		baseConditions = baseConditions.Where("granularity = ?", constants.Daily)

	default:
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

		baseConditions = baseConditions.Select([]string{
			fmt.Sprintf("DATE_TRUNC('%s', period_start) AS period_start", groupByUnit),
			"SUM(total_orders) AS total_orders",
			"SUM(total_amount::numeric) AS total_amount",
			"SUM(total_transferred::numeric) AS total_transferred",
			"symbol",
			"vendor_id",
		}).
			Where("granularity = ?", constants.Daily). // aggregate from daily data
			Group("symbol, vendor_id, DATE_TRUNC('" + groupByUnit + "', period_start)")
	}

	if err := baseConditions.Find(&statistics).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment statistics: %w", err)
	}

	return statistics, nil
}
