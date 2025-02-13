package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type UserWalletRepository interface {
	CreateUserWallets(ctx context.Context, userWallets []entities.UserWallet) error
	GetUserWallets(
		ctx context.Context,
		limit, offset int,
		userIDs []uint64,
	) ([]entities.UserWallet, error)
}

type UserWalletUCase interface {
	CreateUserWallets(ctx context.Context, payloads []dto.UserWalletPayloadDTO) error
	GetUserWallets(
		ctx context.Context,
		page, size int,
		userIDs []string,
	) (dto.PaginationDTOResponse, error)
}
