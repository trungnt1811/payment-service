package block_state

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type blockstateRepository struct {
	db *gorm.DB
}

// NewBlockstateRepository creates a new BlockStateRepository
func NewBlockstateRepository(db *gorm.DB) interfaces.BlockStateRepository {
	return &blockstateRepository{
		db: db,
	}
}

// GetLastProcessedBlock retrieves the last processed block from the database.
func (r *blockstateRepository) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	var blockState model.BlockState

	// Attempt to find the first block state entry in the database.
	if err := r.db.WithContext(ctx).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 0 if no record is found (first run scenario)
			return 0, nil
		}
		return 0, err
	}

	return blockState.LastProcessedBlock, nil
}

// UpdateLastProcessedBlock updates the last processed block in the database.
func (r *blockstateRepository) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64) error {
	var blockState model.BlockState

	// Check if the block state record exists
	if err := r.db.WithContext(ctx).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no record is found, create a new one
			blockState.LastProcessedBlock = blockNumber
			return r.db.WithContext(ctx).Create(&blockState).Error
		}
		return err
	}

	// If record exists, update the last block number
	blockState.LastProcessedBlock = blockNumber
	return r.db.WithContext(ctx).Save(&blockState).Error
}

// GetLatestBlock retrieves the latest block from the database.
func (r *blockstateRepository) GetLatestBlock(ctx context.Context) (uint64, error) {
	var blockState model.BlockState

	// Attempt to find the first block state entry in the database.
	if err := r.db.WithContext(ctx).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return 0 if no record is found (first run scenario)
			return 0, nil
		}
		return 0, err
	}

	return blockState.LatestBlock, nil
}

// UpdateLatestBlock updates the latest block in the database.
func (r *blockstateRepository) UpdateLatestBlock(ctx context.Context, blockNumber uint64) error {
	var blockState model.BlockState

	// Check if the block state record exists
	if err := r.db.WithContext(ctx).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no record is found, create a new one
			blockState.LatestBlock = blockNumber
			return r.db.WithContext(ctx).Create(&blockState).Error
		}
		return err
	}

	// If record exists, update the latest block number
	blockState.LatestBlock = blockNumber
	return r.db.WithContext(ctx).Save(&blockState).Error
}
