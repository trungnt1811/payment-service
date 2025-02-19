package types

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

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
