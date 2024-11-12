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

// GetLastProcessedBlock retrieves the last processed block for the specified network from the database.
func (r *blockstateRepository) GetLastProcessedBlock(ctx context.Context, network string) (uint64, error) {
	var blockState model.BlockState

	// Attempt to find the block state entry for the specified network
	if err := r.db.WithContext(ctx).Where("network = ?", network).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return blockState.LastProcessedBlock, nil
}

// UpdateLastProcessedBlock updates the last processed block for the specified network in the database.
func (r *blockstateRepository) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network string) error {
	var blockState model.BlockState

	// Check if the block state record for the network exists
	if err := r.db.WithContext(ctx).Where("network = ?", network).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new record if none exists for this network
			blockState.Network = network
			blockState.LastProcessedBlock = blockNumber
			return r.db.WithContext(ctx).Create(&blockState).Error
		}
		return err
	}

	// Update the last processed block for the network
	blockState.LastProcessedBlock = blockNumber
	return r.db.WithContext(ctx).Save(&blockState).Error
}

// GetLatestBlock retrieves the latest block for the specified network from the database.
func (r *blockstateRepository) GetLatestBlock(ctx context.Context, network string) (uint64, error) {
	var blockState model.BlockState

	// Attempt to find the block state entry for the specified network
	if err := r.db.WithContext(ctx).Where("network = ?", network).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return blockState.LatestBlock, nil
}

// UpdateLatestBlock updates the latest block for the specified network in the database.
func (r *blockstateRepository) UpdateLatestBlock(ctx context.Context, blockNumber uint64, network string) error {
	var blockState model.BlockState

	// Check if the block state record for the network exists
	if err := r.db.WithContext(ctx).Where("network = ?", network).First(&blockState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new record if none exists for this network
			blockState.Network = network
			blockState.LatestBlock = blockNumber
			return r.db.WithContext(ctx).Create(&blockState).Error
		}
		return err
	}

	// Update the latest block for the network
	blockState.LatestBlock = blockNumber
	return r.db.WithContext(ctx).Save(&blockState).Error
}
