package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type TokenTransferRepository interface {
	CreateTokenTransferHistories(ctx context.Context, models []model.TokenTransferHistory) error
	GetTokenTransferHistories(ctx context.Context, filters map[string]interface{}, page, size int) ([]model.TokenTransferHistory, error)
}

type TokenTransferUCase interface {
	TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) error
	GetTokenTransferHistories(ctx context.Context, filters dto.TokenTransferFilterDTO, page, size int) (dto.TokenTransferHistoryDTOResponse, error)
}
