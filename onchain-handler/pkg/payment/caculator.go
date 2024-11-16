package payment

import "math/big"

func CalculatePaymentCovering(amount *big.Int, paymentCoveringFactor float64) *big.Int {
	// Convert the covering factor to a multiplier as a float64.
	coveringFactorMultiplier := 1 - (paymentCoveringFactor / 100)

	amountFloat := new(big.Float).SetInt(amount)
	coveringFactorFloat := big.NewFloat(coveringFactorMultiplier)

	// Perform the multiplication (big.Float * big.Float)
	minimumAcceptedAmountFloat := new(big.Float).Mul(amountFloat, coveringFactorFloat)

	// Convert the result back to a big.Int (this rounds the float result)
	minimumAcceptedAmount := new(big.Int)
	minimumAcceptedAmountFloat.Int(minimumAcceptedAmount)
	return minimumAcceptedAmount
}
