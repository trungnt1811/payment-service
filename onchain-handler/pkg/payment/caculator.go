package payment

import "math/big"

// CalculatePaymentCovering calculates the minimum accepted payment amount based on a given amount and payment covering factor.
// amount: The original payment amount as *big.Int.
// paymentCoveringFactor: The covering factor as a percentage (e.g., 1.2 for 1.2%).
// Returns: The minimum accepted amount as *big.Int.
func CalculatePaymentCovering(amount *big.Int, paymentCoveringFactor float64) *big.Int {
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
