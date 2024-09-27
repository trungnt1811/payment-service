package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type LockRepository interface {
	CreateLockEventHistory(ctx context.Context, model model.LockEvent) error
	GetLatestLockEventsByUserAddress(ctx context.Context, userAddress string, page, size int) ([]model.LockEvent, error)
	GetDepositLockEventByLockIDs(ctx context.Context, lockIDs []uint64) ([]model.LockEvent, error)
	GetLockEventHistoriesByUserAddress(ctx context.Context, userAddress string, page, size int) ([]model.LockEvent, error)
}

type LockUCase interface {
	GetLatestLockEventsByUserAddress(ctx context.Context, userAddress string, page, size int) (dto.LockEventDTOResponse, error)
}
