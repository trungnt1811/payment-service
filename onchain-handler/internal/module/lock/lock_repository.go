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
