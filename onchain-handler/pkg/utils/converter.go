package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// ConvertFloatTokenToSmallestUnit converts a float string amount in Eth (or any token) into its equivalent in Wei (or smallest unit) (big.Int).
func ConvertFloatTokenToSmallestUnit(amount string, decimals int) (*big.Int, error) {
	// Split the amount into integer and fractional parts.
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid format: multiple decimal points")
	}

	// Convert the integer part to big.Int.
	intPart := new(big.Int)
	_, success := intPart.SetString(parts[0], 10)
	if !success {
		return nil, fmt.Errorf("invalid integer part")
	}

	// Initialize multiplier for conversion (10^decimals).
	weiMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// If there's no fractional part, simply multiply the integer part by the multiplier.
	if len(parts) == 1 {
		return new(big.Int).Mul(intPart, weiMultiplier), nil
	}

	// Handle the fractional part.
	fractionPart := parts[1]

	// Check if the fractional part exceeds the allowed decimal places.
	if len(fractionPart) > decimals {
		return nil, fmt.Errorf("fractional part exceeds %d decimal places", decimals)
	}

	// Pad the fractional part to the required decimal places by appending zeros.
	fractionPartPadded := fractionPart + strings.Repeat("0", decimals-len(fractionPart))

	// Convert the fractional part to big.Int.
	fracPart := new(big.Int)
	_, success = fracPart.SetString(fractionPartPadded, 10)
	if !success {
		return nil, fmt.Errorf("invalid fractional part")
	}

	// Multiply the integer part by the multiplier and add the fractional part.
	result := new(big.Int).Mul(intPart, weiMultiplier)
	result.Add(result, fracPart)

	return result, nil
}

// ConvertSmallestUnitToFloatToken converts a smallest unit (e.g., Wei) into its float representation (e.g., Ether).
func ConvertSmallestUnitToFloatToken(amount string, decimals int) (string, error) {
	// Parse the input amount as a big.Int.
	weiAmount := new(big.Int)
	_, ok := weiAmount.SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid wei amount format")
	}

	// Convert the Wei amount to a big.Float for division.
	weiFloat := new(big.Float).SetInt(weiAmount)

	// Calculate the divisor for the given decimals (10^decimals).
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// Divide the Wei amount by the divisor to get the Eth amount.
	ethFloat := new(big.Float).Quo(weiFloat, divisor)

	// Convert the result to a string with full precision (up to decimals places).
	ethValue := ethFloat.Text('f', decimals)

	return ethValue, nil
}
