package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type TransferRepository interface {
	CreateTransferHistories(ctx context.Context, models []model.TransferHistory) error
}

type TransferUCase interface {
	DistributeTokens(ctx context.Context, payloads []dto.TransferTokenPayloadDTO) error
}
