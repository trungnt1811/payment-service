package user_wallet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type userWalletUCase struct {
	userWalletRepository interfaces.UserWalletRepository
}

func NewUserWalletUCase(
	userWalletRepository interfaces.UserWalletRepository,
) interfaces.UserWalletUCase {
	return &userWalletUCase{
		userWalletRepository: userWalletRepository,
	}
}

func (u *userWalletUCase) CreateUserWallets(ctx context.Context, payloads []dto.UserWalletPayloadDTO) error {
	var wallets []model.UserWallet
	for _, payload := range payloads {
		wallet := model.UserWallet{
			UserID:  payload.UserID,
			Address: payload.Address,
		}
		wallets = append(wallets, wallet)
	}
	return u.userWalletRepository.CreateUserWallets(ctx, wallets)
}

func (u *userWalletUCase) GetUserWallets(
	ctx context.Context,
	page, size int,
	userIDs []string,
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	// Convert userIDs from strings to uint64s
	var userIDUInts []uint64
	for _, idStr := range userIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return dto.PaginationDTOResponse{}, fmt.Errorf("invalid user ID '%s': %v", idStr, err)
		}
		userIDUInts = append(userIDUInts, id)
	}

	// Fetch the wallets from the repository
	wallets, err := u.userWalletRepository.GetUserWallets(ctx, limit, offset, userIDUInts)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	// Map wallets to DTOs, limiting to requested page size
	var walletDTOs []interface{}
	for i, wallet := range wallets {
		if i >= size { // Stop if we reach the requested page size
			break
		}
		walletDTOs = append(walletDTOs, wallet.ToDto())
	}

	// Determine if there's a next page
	nextPage := page
	if len(wallets) > size {
		nextPage += 1
	}

	// Return the response DTO
	return dto.PaginationDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     walletDTOs,
	}, nil
}
