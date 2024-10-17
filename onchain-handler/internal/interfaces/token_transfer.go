package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type TokenTransferRepository interface {
	CreateTokenTransferHistories(ctx context.Context, models []model.TokenTransferHistory) error
	GetTokenTransferHistories(
		ctx context.Context,
		limit, offset int,
		requestIDs []string,
		startTime, endTime time.Time,
	) ([]model.TokenTransferHistory, error)
}

type TokenTransferUCase interface {
	TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) ([]dto.TokenTransferResultDTOResponse, error)
	GetTokenTransferHistories(
		ctx context.Context,
		requestIDs []string,
		startTime, endTime time.Time,
		page, size int,
	) (dto.PaginationDTOResponse, error)
}
