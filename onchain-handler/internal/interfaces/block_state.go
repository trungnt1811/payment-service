package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
)

type BlockStateRepository interface {
	GetLastProcessedBlock(ctx context.Context, network string) (uint64, error)
	UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network string) error
	GetLatestBlock(ctx context.Context, network string) (uint64, error)
	UpdateLatestBlock(ctx context.Context, blockNumber uint64, network string) error
}

type BlockStateUCase interface {
	GetLastProcessedBlock(ctx context.Context, network constants.NetworkType) (uint64, error)
	UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error
	GetLatestBlock(ctx context.Context, network constants.NetworkType) (uint64, error)
	UpdateLatestBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error
}
