package types

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type TokenTransferRepository interface {
	CreateTokenTransferHistories(ctx context.Context, models []entities.TokenTransferHistory) error
	GetTokenTransferHistories(
		ctx context.Context,
		limit, offset int,
		orderBy *string,
		orderDirection constants.OrderDirection,
		startTime, endTime *time.Time,
		fromAddress, toAddress *string,
	) ([]entities.TokenTransferHistory, error)
	GetTotalTokenAmount(
		ctx context.Context,
		startTime, endTime *time.Time,
		fromAddress, toAddress *string,
	) (float64, error)
}
