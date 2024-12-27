package payment

import (
	"math/big"
	"strconv"

	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// CalculatePaymentCovering calculates the minimum accepted payment amount based on a given amount and payment covering factor.
// amount: The original payment amount as *big.Int.
// paymentCoveringFactor: The covering factor as a percentage (e.g., 1.2 for 1.2%).
// Returns: The minimum accepted amount as *big.Int.
func CalculatePaymentCoveringAsPercentage(amount *big.Int, paymentCoveringFactor float64) *big.Int {
	// Validate payment covering factor to avoid invalid calculations.
	if paymentCoveringFactor < 0 || paymentCoveringFactor > 100 {
		return new(big.Int).Set(amount) // Return the original amount as a fallback.
	}

	// Calculate the multiplier: (1 - (paymentCoveringFactor / 100)).
	coveringFactorMultiplier := 1 - (paymentCoveringFactor / 100)

	// Convert amount to big.Float for precise floating-point calculations.
	amountFloat := new(big.Float).SetInt(amount)

	// Convert multiplier to big.Float for compatibility.
	coveringFactorFloat := big.NewFloat(coveringFactorMultiplier)

	// Perform the multiplication: amount * coveringFactorMultiplier.
	minimumAcceptedAmountFloat := new(big.Float).Mul(amountFloat, coveringFactorFloat)

	// Convert the resulting big.Float back to a big.Int (rounding down the result).
	minimumAcceptedAmount := new(big.Int)
	minimumAcceptedAmountFloat.Int(minimumAcceptedAmount)

	return minimumAcceptedAmount
}

// CalculatePaymentCoveringAsDiscount calculates the payment amount after applying a fixed covering (discount).
// amount: The original payment amount as *big.Int.
// paymentCoveringFactor: The covering factor as a fixed value (e.g., 1 for a discount of 1 USDT).
// Returns: The discounted amount as *big.Int.
func CalculatePaymentCoveringAsDiscount(amount *big.Int, paymentCoveringFactor float64, decimals uint8) *big.Int {
	discount, err := utils.ConvertFloatTokenToSmallestUnit(strconv.FormatFloat(paymentCoveringFactor, 'f', -1, 64), decimals)
	if err != nil {
		// If conversion fails, log and return the original amount as fallback.
		logger.GetLogger().Errorf("Failed to convert payment covering factor to smallest unit: %v", err)
		return new(big.Int).Set(amount)
	}

	// Subtract the discount from the original amount.
	discountedAmount := new(big.Int).Sub(amount, discount)

	// Ensure the discounted amount is not less than 0.
	if discountedAmount.Sign() < 0 {
		return big.NewInt(0)
	}

	return discountedAmount
}
