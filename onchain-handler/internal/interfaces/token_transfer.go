package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type TokenTransferRepository interface {
	CreateTokenTransferHistories(ctx context.Context, models []model.TokenTransferHistory) error
}

type TokenTransferUCase interface {
	TransferTokens(ctx context.Context, payloads []dto.TransferTokenPayloadDTO) error
}
