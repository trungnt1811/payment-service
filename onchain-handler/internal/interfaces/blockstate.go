package interfaces

import "context"

type BlockStateRepository interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64) error
}
