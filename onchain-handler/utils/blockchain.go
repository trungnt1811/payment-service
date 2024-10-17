package utils

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/constants"
)

// PollForLogsFromBlock polls logs from a specified block number onwards for the given contract.
func PollForLogsFromBlock(
	ctx context.Context,
	client *ethclient.Client, // Ethereum client
	contractAddresses []common.Address, // Contract addresses to filter logs
	fromBlock uint64, // Block number to start querying from
	endBlock uint64,
) ([]types.Log, error) {
	// Prepare filter query for the logs
	query := ethereum.FilterQuery{
		Addresses: contractAddresses,            // Contract addresses to filter logs from
		FromBlock: big.NewInt(int64(fromBlock)), // Start block for querying logs
		ToBlock:   big.NewInt(int64(endBlock)),  // End block for querying logs
	}

	// Poll for logs
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs from block %d: %w", fromBlock, err)
	}

	return logs, nil
}

// GetLatestBlockNumber retrieves the latest block number from the Ethereum client
func GetLatestBlockNumber(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest block header: %w", err)
	}
	return header.Number, nil
}

// ConvertFloatAvaxToWei converts a float string amount in AVAX into its equivalent in Wei (big.Int).
func ConvertFloatAvaxToWei(amount string) (*big.Int, error) {
	// Split the amount into integer and fractional parts.
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid float avax amount format")
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

func ConvertWeiToAvax(amount string) (string, error) {
	weiAmount := new(big.Int)
	_, ok := weiAmount.SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid wei amount format")
	}

	// Use big.Float to handle the conversion from Wei to AVAX (1 AVAX = 10^18 Wei)
	avaxAmount := new(big.Float).SetInt(weiAmount)
	avaxValue := new(big.Float).Quo(avaxAmount, big.NewFloat(constants.TokenDecimalsMultiplier)) // Divide by 10^18

	// Convert the result to a string for saving (to avoid loss of precision)
	amountInAvax := avaxValue.Text('f', 18) // 18 decimal places for precision in AVAX
	return amountInAvax, nil
}
