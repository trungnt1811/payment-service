package lock

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type lockRepository struct {
	db *gorm.DB
}

// NewLockRepository creates a new new LockRepository
func NewLockRepository(db *gorm.DB) interfaces.LockRepository {
	return &lockRepository{
		db: db,
	}
}

// CreateLockEventHistory stores the lock event in the database.
func (r *lockRepository) CreateLockEventHistory(ctx context.Context, lockEvent model.LockEvent) error {
	if err := r.db.WithContext(ctx).Create(&lockEvent).Error; err != nil {
		return err
	}
	return nil
}

// GetLatestLockEventsByUserAddress retrieves the latest lock event for each lock ID of a user.
func (r *lockRepository) GetLatestLockEventsByUserAddress(ctx context.Context, userAddress string) ([]model.LockEvent, error) {
	var lockEvents []model.LockEvent

	if err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (lock_id) * 
			FROM lock_event 
			WHERE user_address = ? AND status = 1
			ORDER BY lock_id, created_at DESC
		`, userAddress).Scan(&lockEvents).Error; err != nil {
		return nil, err
	}

	return lockEvents, nil
}

// GetDepositLockEventByLockIDs retrieves deposit lock events for the provided lock IDs.
func (r *lockRepository) GetDepositLockEventByLockIDs(ctx context.Context, lockIDs []uint64) ([]model.LockEvent, error) {
	var lockEvents []model.LockEvent

	if err := r.db.WithContext(ctx).
		Where("lock_id IN ? AND lock_action = ?", lockIDs, "DEPOSIT").
		Find(&lockEvents).Error; err != nil {
		return nil, err
	}

	return lockEvents, nil
}

// GetLockEventHistoriesByUserAddress retrieves (both deposit and withdraw) lock events for a specific user.
func (r *lockRepository) GetLockEventHistoriesByUserAddress(ctx context.Context, userAddress string) ([]model.LockEvent, error) {
	var lockEvents []model.LockEvent

	if err := r.db.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Find(&lockEvents).Error; err != nil {
		return nil, err
	}

	return lockEvents, nil
}
