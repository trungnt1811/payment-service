package types

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
	GetTokenDecimals(ctx context.Context, tokenContractAddress string) (uint8, error)
	EstimateGasGeneric(
		contractAddress common.Address,
		fromAddress common.Address,
		abiDef string,
		method string,
		args ...interface{},
	) (uint64, error)
	TransferToken(
		ctx context.Context,
		chainID uint64,
		tokenContractAddress, fromPrivateKeyHex, toAddressHex string,
		amount *big.Int,
	) (common.Hash, uint64, *big.Int, uint64, error)
	TransferNativeToken(
		ctx context.Context,
		chainID uint64,
		fromPrivateKeyHex, toAddressHex string,
		amount *big.Int,
	) (common.Hash, uint64, *big.Int, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	GetBaseFee(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	GetTokenBalance(
		ctx context.Context,
		tokenContractAddress string,
		walletAddress string,
	) (*big.Int, error)
	GetNativeTokenBalance(
		ctx context.Context,
		walletAddress string,
	) (*big.Int, error)
	Close()
}
