package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// InitClient init eth client
func InitClient(rpcUrl string) (*ethclient.Client, error) {
	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	return ethClient, nil
}

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
