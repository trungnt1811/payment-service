package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
)

type BlockStateUCase interface {
	GetLastProcessedBlock(ctx context.Context, network constants.NetworkType) (uint64, error)
	UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error
	GetLatestBlock(ctx context.Context, network constants.NetworkType) (uint64, error)
	UpdateLatestBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error
}
