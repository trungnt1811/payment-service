package entities

import (
	"sort"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

// PaymentStatistics represents the payment statistics domain model.
type PaymentStatistics struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Granularity      string    `json:"granularity"`  // Enum: DAILY, WEEKLY, MONTHLY, YEARLY
	PeriodStart      time.Time `json:"period_start"` // Start of the period
	TotalOrders      uint64    `json:"total_orders"`
	TotalAmount      string    `json:"total_amount"`
	TotalTransferred string    `json:"total_transferred"`
	Symbol           string    `json:"symbol"`
	VendorID         string    `json:"vendor_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName overrides the default table name for GORM.
func (m *PaymentStatistics) TableName() string {
	return "payment_statistics"
}

// ToPeriodStatisticsDTO converts a slice of PaymentStatistics to a slice of PeriodStatistics DTOs.
func ToPeriodStatisticsDTO(data []PaymentStatistics) []dto.PeriodStatistics {
	grouped := make(map[uint64]map[string]dto.TokenStats)

	// Build grouped map: map[periodStart] => map[symbol] => TokenStats
	for _, stat := range data {
		periodStart := uint64(stat.PeriodStart.UTC().Unix())

		if _, ok := grouped[periodStart]; !ok {
			grouped[periodStart] = make(map[string]dto.TokenStats)
		}

		amount := stat.TotalAmount
		if amount == "" {
			amount = "0"
		}

		transferred := stat.TotalTransferred
		if transferred == "" {
			transferred = "0"
		}

		grouped[periodStart][stat.Symbol] = dto.TokenStats{
			Symbol:           stat.Symbol,
			TotalOrders:      stat.TotalOrders,
			TotalAmount:      amount,
			TotalTransferred: transferred,
		}
	}

	// Prepare and sort output
	var sortedKeys []uint64
	for k := range grouped {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] > sortedKeys[j] // DESC
	})

	var result []dto.PeriodStatistics
	for _, ts := range sortedKeys {
		tokenMap := grouped[ts]

		// Ensure both USDC and USDT exist
		if _, ok := tokenMap[constants.USDC]; !ok {
			tokenMap[constants.USDC] = dto.TokenStats{
				Symbol:           constants.USDC,
				TotalOrders:      0,
				TotalAmount:      "0",
				TotalTransferred: "0",
			}
		}
		if _, ok := tokenMap[constants.USDT]; !ok {
			tokenMap[constants.USDT] = dto.TokenStats{
				Symbol:           constants.USDT,
				TotalOrders:      0,
				TotalAmount:      "0",
				TotalTransferred: "0",
			}
		}

		// Convert to slice with USDC always first
		stats := []dto.TokenStats{
			tokenMap[constants.USDC],
			tokenMap[constants.USDT],
		}

		result = append(result, dto.PeriodStatistics{
			PeriodStart: ts,
			TokenStats:  stats,
		})
	}

	return result
}
