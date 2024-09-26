package membership

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type membershipRepository struct {
	db *gorm.DB
}

func NewMembershipRepository(db *gorm.DB) interfaces.MembershipRepository {
	return &membershipRepository{
		db: db,
	}
}

func (r *membershipRepository) CreateMembershipEventHistory(ctx context.Context, membershipEvent model.MembershipEvent) error {
	if err := r.db.WithContext(ctx).Create(&membershipEvent).Error; err != nil {
		return err
	}
	return nil
}

func (r *membershipRepository) GetMembershipEventByOrderID(ctx context.Context, orderID uint64) (*model.MembershipEvent, error) {
	var membershipEvent model.MembershipEvent
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&membershipEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &membershipEvent, nil
}

// GetMembershipEventsByOrderIDs retrieves a list of membership events based on a slice of order IDs.
func (r *membershipRepository) GetMembershipEventsByOrderIDs(ctx context.Context, orderIDs []uint64) ([]model.MembershipEvent, error) {
	var membershipEvents []model.MembershipEvent
	if len(orderIDs) == 0 {
		return nil, fmt.Errorf("orderIDs cannot be empty")
	}

	// Query the database to get all events matching the given order IDs.
	if err := r.db.WithContext(ctx).
		Where("order_id IN ?", orderIDs).
		Find(&membershipEvents).Error; err != nil {
		return nil, err
	}

	return membershipEvents, nil
}
