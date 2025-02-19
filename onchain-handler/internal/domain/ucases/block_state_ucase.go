package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

type blockStateUCase struct {
	blockStateRepo repotypes.BlockStateRepository
}

func NewBlockStateUCase(blockStateRepo repotypes.BlockStateRepository) ucasetypes.BlockStateUCase {
	return &blockStateUCase{
		blockStateRepo: blockStateRepo,
	}
}

func (u *blockStateUCase) GetLatestBlock(ctx context.Context, network constants.NetworkType) (uint64, error) {
	return u.blockStateRepo.GetLatestBlock(ctx, network.String())
}

func (u *blockStateUCase) UpdateLatestBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error {
	return u.blockStateRepo.UpdateLatestBlock(ctx, blockNumber, network.String())
}

func (u *blockStateUCase) GetLastProcessedBlock(ctx context.Context, network constants.NetworkType) (uint64, error) {
	return u.blockStateRepo.GetLastProcessedBlock(ctx, network.String())
}

func (u *blockStateUCase) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network constants.NetworkType) error {
	return u.blockStateRepo.UpdateLastProcessedBlock(ctx, blockNumber, network.String())
}
