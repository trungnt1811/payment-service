package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type UserWalletRepository interface {
	CreateUserWallets(ctx context.Context, userWallets []domain.UserWallet) error
	GetUserWallets(
		ctx context.Context,
		limit, offset int,
		userIDs []uint64,
	) ([]domain.UserWallet, error)
}

type UserWalletUCase interface {
	CreateUserWallets(ctx context.Context, payloads []dto.UserWalletPayloadDTO) error
	GetUserWallets(
		ctx context.Context,
		page, size int,
		userIDs []string,
	) (dto.PaginationDTOResponse, error)
}
