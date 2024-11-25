package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Client interface {
	PollForLogsFromBlock(
		ctx context.Context,
		contractAddresses []common.Address, // Contract addresses to filter logs
		fromBlock uint64, // Block number to start querying from
		endBlock uint64,
	) ([]types.Log, error)
	GetLatestBlockNumber(ctx context.Context) (*big.Int, error)
	BulkTransfer(
		ctx context.Context,
		chainID uint64,
		bulkSenderContractAddress, poolAddress, poolPrivateKey, tokenContractAddress string,
		recipients []string,
		amounts []*big.Int,
	) (*string, *string, *big.Float, error)
	GetTokenDecimals(ctx context.Context, tokenContractAddress string) (uint8, error)
	Close()
}
