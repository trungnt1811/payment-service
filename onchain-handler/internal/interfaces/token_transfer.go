package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
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

type TokenTransferUCase interface {
	CreateTokenTransferHistories(ctx context.Context, payloads []dto.TokenTransferHistoryDTO) error
	GetTokenTransferHistories(
		ctx context.Context,
		startTime, endTime *time.Time,
		orderBy *string,
		orderDirection constants.OrderDirection,
		page, size int,
		fromAddress, toAddress *string,
	) (dto.PaginationDTOResponse, error)
	GetTotalTokenAmount(
		ctx context.Context,
		startTime, endTime *time.Time,
		fromAddress, toAddress *string,
	) (float64, error)
}
