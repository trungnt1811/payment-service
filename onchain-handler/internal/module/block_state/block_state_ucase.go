package block_state

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type blockStateUCase struct {
	blockStateRepo interfaces.BlockStateRepository
}

func NewBlockStateUCase(blockStateRepo interfaces.BlockStateRepository) interfaces.BlockStateUCase {
	return &blockStateUCase{
		blockStateRepo: blockStateRepo,
	}
}

func (u *blockStateUCase) GetLatestBlock(ctx context.Context, network constants.NetworkType) (uint64, error) {
	return u.blockStateRepo.GetLatestBlock(ctx, string(network))
}

func (u *blockStateUCase) UpdateLatestBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error {
	return u.blockStateRepo.UpdateLatestBlock(ctx, blockNumber, string(network))
}

func (u *blockStateUCase) GetLastProcessedBlock(ctx context.Context, network constants.NetworkType) (uint64, error) {
	return u.blockStateRepo.GetLastProcessedBlock(ctx, string(network))
}

func (u *blockStateUCase) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error {
	return u.blockStateRepo.UpdateLastProcessedBlock(ctx, blockNumber, string(network))
}
