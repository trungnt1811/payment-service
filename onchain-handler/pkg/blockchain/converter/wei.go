package converter

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/genefriendway/onchain-handler/constants"
)

// ConvertFloatEthToWei converts a float string amount in Eth into its equivalent in Wei (big.Int).
func ConvertFloatEthToWei(amount string) (*big.Int, error) {
	// Split the amount into integer and fractional parts.
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid float Eth amount format")
	}

	// Convert the integer part to big.Int.
	intPart := new(big.Int)
	_, success := intPart.SetString(parts[0], 10)
	if !success {
		return nil, fmt.Errorf("invalid integer part")
	}

	// Initialize multiplier for Wei conversion (10^18).
	weiMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(constants.DecimalPlaces), nil)

	// If there's no fractional part, simply multiply the integer part by 10^18.
	if len(parts) == 1 {
		return intPart.Mul(intPart, weiMultiplier), nil
	}

	// If there's a fractional part, handle it.
	fractionPart := parts[1]

	// Check if the fractional part exceeds 18 digits.
	if len(fractionPart) > constants.DecimalPlaces {
		return nil, fmt.Errorf("fractional part too long")
	}

	// Pad the fractional part to 18 digits by appending zeros.
	fractionPartPadded := fractionPart + strings.Repeat("0", constants.DecimalPlaces-len(fractionPart))

	// Convert the fractional part to big.Int.
	fracPart := new(big.Int)
	_, success = fracPart.SetString(fractionPartPadded, 10)
	if !success {
		return nil, fmt.Errorf("invalid fractional part")
	}

	// Multiply integer part by 10^18 and add the fractional part.
	result := new(big.Int).Mul(intPart, weiMultiplier)
	result.Add(result, fracPart)

	return result, nil
}

func ConvertWeiToEth(amount string) (string, error) {
	weiAmount := new(big.Int)
	_, ok := weiAmount.SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid wei amount format")
	}

	// Use big.Float to handle the conversion from Wei to Eth (1 Eth = 10^18 Wei)
	ethAmount := new(big.Float).SetInt(weiAmount)
	ethValue := new(big.Float).Quo(ethAmount, big.NewFloat(constants.TokenDecimalsMultiplier)) // Divide by 10^18

	// Convert the result to a string for saving (to avoid loss of precision)
	amountInEth := ethValue.Text('f', 18) // 18 decimal places for precision in Eth
	return amountInEth, nil
}
