package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type UserWalletRepository interface {
	CreateUserWallets(ctx context.Context, userWallets []model.UserWallet) error
	GetUserWallets(
		ctx context.Context,
		limit, offset int,
		userIDs []uint64,
	) ([]model.UserWallet, error)
}

type UserWalletUCase interface {
	CreateUserWallets(ctx context.Context, payloads []dto.UserWalletPayloadDTO) error
	GetUserWallets(
		ctx context.Context,
		page, size int,
		userIDs []string,
	) (dto.PaginationDTOResponse, error)
}
