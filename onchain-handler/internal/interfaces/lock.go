package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/model"
)

type LockRepository interface {
	CreateLockEventHistory(ctx context.Context, model model.LockEvent) error
}
