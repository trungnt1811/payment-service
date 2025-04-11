package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type blockstateRepository struct {
	db *gorm.DB
}

// NewBlockStateRepository creates a new BlockStateRepository
func NewBlockStateRepository(db *gorm.DB) repotypes.BlockStateRepository {
	return &blockstateRepository{
		db: db,
	}
}

// GetLatestBlock retrieves the latest block with a lock to prevent race conditions
func (r *blockstateRepository) GetLatestBlock(ctx context.Context, network string) (uint64, error) {
	var latestBlock uint64
	err := r.db.WithContext(ctx).
		Raw("SELECT latest_block FROM block_state WHERE network = ?", network).
		Scan(&latestBlock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return latestBlock, nil
}

// UpdateLatestBlock updates only the latest_block column efficiently
func (r *blockstateRepository) UpdateLatestBlock(ctx context.Context, blockNumber uint64, network string) error {
	result := r.db.WithContext(ctx).
		Model(&entities.BlockState{}).
		Where("network = ?", network).
		Updates(map[string]any{"latest_block": blockNumber})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Create new one if not record updated
		blockState := entities.BlockState{
			Network:     network,
			LatestBlock: blockNumber,
		}
		return r.db.WithContext(ctx).Create(&blockState).Error
	}

	return nil
}

func (r *blockstateRepository) GetLastProcessedBlock(ctx context.Context, network string) (uint64, error) {
	var lastProcessedBlock uint64
	err := r.db.WithContext(ctx).
		Raw("SELECT last_processed_block FROM block_state WHERE network = ?", network).
		Scan(&lastProcessedBlock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return lastProcessedBlock, nil
}

// UpdateLastProcessedBlock updates only the last_processed_block column efficiently
func (r *blockstateRepository) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64, network string) error {
	result := r.db.WithContext(ctx).
		Model(&entities.BlockState{}).
		Where("network = ?", network).
		Updates(map[string]any{"last_processed_block": blockNumber})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Create new one if not record updated
		blockState := entities.BlockState{
			Network:            network,
			LastProcessedBlock: blockNumber,
		}
		return r.db.WithContext(ctx).Create(&blockState).Error
	}

	return nil
}
