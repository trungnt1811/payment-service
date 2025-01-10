package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type TokenTransferRepository interface {
	CreateTokenTransferHistories(ctx context.Context, models []domain.TokenTransferHistory) error
	GetTokenTransferHistories(
		ctx context.Context,
		limit, offset int,
		orderBy *string,
		orderDirection constants.OrderDirection,
		startTime, endTime *time.Time,
		fromAddress, toAddress *string,
	) ([]domain.TokenTransferHistory, error)
}

type TokenTransferUCase interface {
	CreateTokenTransferHistories(ctx context.Context, payloads []dto.TokenTransferHistoryDTO) error
	TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) ([]dto.TokenTransferResultDTOResponse, error)
	GetTokenTransferHistories(
		ctx context.Context,
		startTime, endTime *time.Time,
		orderBy *string,
		orderDirection constants.OrderDirection,
		page, size int,
		fromAddress, toAddress *string,
	) (dto.PaginationDTOResponse, error)
}
