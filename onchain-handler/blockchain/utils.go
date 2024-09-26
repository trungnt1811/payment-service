package blockchain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgconn"
)

// loadABI loads and parses the ABI from a JSON file
func loadABI(path string) (abi.ABI, error) {
	file, err := os.Open(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to open ABI file at %s: %w", path, err)
	}
	defer file.Close()

	// Read and parse the ABI JSON file
	var parsedABI abi.ABI
	if err := json.NewDecoder(file).Decode(&parsedABI); err != nil {
		return abi.ABI{}, fmt.Errorf("failed to decode ABI JSON from file %s: %w", path, err)
	}

	return parsedABI, nil
}

// pollForLogsFromBlock polls logs from a specified block number onwards for the given contract.
func pollForLogsFromBlock(
	ctx context.Context,
	client *ethclient.Client, // Ethereum client
	contractAddr common.Address, // Contract address to filter logs
	fromBlock uint64, // Block number to start querying from
	endBlock uint64,
) ([]types.Log, error) {
	// Prepare filter query for the logs
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr}, // Contract address to filter logs from
		FromBlock: big.NewInt(int64(fromBlock)),   // Start block for querying logs
		ToBlock:   big.NewInt(int64(endBlock)),    // End block for querying logs
	}

	// Poll for logs
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs from block %d: %w", fromBlock, err)
	}

	return logs, nil
}

// getLatestBlockNumber retrieves the latest block number from the Ethereum client
func getLatestBlockNumber(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest block header: %w", err)
	}
	return header.Number, nil
}

// parseHexToUint64 parses a hex string to uint64.
// Handles hex strings that start with "0x" and ignores leading zeros.
func parseHexToUint64(hexStr string) (uint64, error) {
	// Ensure the string is lowercase and remove the "0x" prefix if present.
	hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")

	// If the string is empty or all zeros, return 0.
	if len(hexStr) == 0 || strings.TrimLeft(hexStr, "0") == "" {
		return 0, nil
	}

	// Decode the hex string into bytes.
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("invalid hex string: %w", err)
	}

	// Convert the bytes to uint64. Only the last 8 bytes are relevant for uint64.
	var value uint64
	for _, b := range bytes {
		value = value<<8 + uint64(b)
	}

	return value, nil
}

// isDuplicateTransactionError checks if the error is due to a unique constraint violation (e.g., duplicate transaction hash).
func isDuplicateTransactionError(err error) bool {
	var pqErr *pgconn.PgError
	// Check if the error is a PostgreSQL error and has a unique violation code
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		// Optionally, further verify if the constraint name is "unique_transaction_hash"
		if strings.Contains(pqErr.Message, "unique_transaction_hash") {
			return true
		}
	}
	return false
}
