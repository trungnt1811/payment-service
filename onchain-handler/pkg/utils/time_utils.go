package utils

import (
	"time"

	"github.com/genefriendway/onchain-handler/constants"
)

// GetPeriodStart calculates the period start time in UTC based on granularity.
// Supported granularities: DAILY, WEEKLY, MONTHLY, YEARLY.
func GetPeriodStart(granularity string, t time.Time) time.Time {
	t = t.UTC() // Normalize to UTC for consistency

	switch granularity {
	case constants.Daily:
		// Start of the current day
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	case constants.Weekly:
		// Start of the current week (Monday)
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7 // Treat Sunday as the 7th day
		}
		return t.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)

	case constants.Monthly:
		// Start of the current month
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)

	case constants.Yearly:
		// Start of the current year
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	default:
		// Default to daily granularity
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
}
