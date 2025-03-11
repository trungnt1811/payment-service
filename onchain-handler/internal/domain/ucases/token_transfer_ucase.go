package ucases

import (
	"context"
	"fmt"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type tokenTransferUCase struct {
	tokenTransferRepository repotypes.TokenTransferRepository
}

func NewTokenTransferUCase(
	tokenTransferRepository repotypes.TokenTransferRepository,
) ucasetypes.TokenTransferUCase {
	return &tokenTransferUCase{
		tokenTransferRepository: tokenTransferRepository,
	}
}

// GetTokenTransferHistories fetches token transfer histories from the repository with optional filters
func (u *tokenTransferUCase) GetTokenTransferHistories(
	ctx context.Context,
	startTime, endTime *time.Time, // Time range to filter by
	orderBy *string, // Order by field
	orderDirection constants.OrderDirection, // Order direction for sorting
	page, size int,
	fromAddress, toAddress *string, // Address filters
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1
	offset := (page - 1) * size

	// Fetch the token transfer histories using the repository with filters
	listTokenTransfers, err := u.tokenTransferRepository.GetTokenTransferHistories(
		ctx, limit, offset, orderBy, orderDirection, startTime, endTime, fromAddress, toAddress,
	)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	var listTokenTransfersDTO []any

	// Convert the token transfer histories to DTO format
	for i := range listTokenTransfers {
		if i >= size {
			break
		}
		listTokenTransfersDTO = append(listTokenTransfersDTO, listTokenTransfers[i].ToDto())
	}

	// Determine the next page if there are more token transfers
	nextPage := page
	if len(listTokenTransfers) > size {
		nextPage += 1
	}

	// Return the response DTO
	return dto.PaginationDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     listTokenTransfersDTO,
	}, nil
}

func (u *tokenTransferUCase) CreateTokenTransferHistories(ctx context.Context, payloads []dto.TokenTransferHistoryDTO) error {
	var models []entities.TokenTransferHistory

	for _, payload := range payloads {
		models = append(models, entities.TokenTransferHistory{
			Network:         payload.Network,
			TransactionHash: payload.TransactionHash,
			FromAddress:     payload.FromAddress,
			ToAddress:       payload.ToAddress,
			TokenAmount:     payload.TokenAmount,
			Fee:             payload.Fee,
			Symbol:          payload.Symbol,
			Status:          payload.Status,
			ErrorMessage:    payload.ErrorMessage,
			Type:            payload.Type,
		})
	}

	return u.tokenTransferRepository.CreateTokenTransferHistories(ctx, models)
}

// GetTotalTokenAmount retrieves the total token amount for the specified filters.
func (u *tokenTransferUCase) GetTotalTokenAmount(
	ctx context.Context,
	startTime, endTime *time.Time,
	fromAddress, toAddress *string,
) (float64, error) {
	// Call the repository method to calculate the total token amount
	totalTokenAmount, err := u.tokenTransferRepository.GetTotalTokenAmount(ctx, startTime, endTime, fromAddress, toAddress)
	if err != nil {
		logger.GetLogger().Errorf("Failed to calculate total token amount: %v", err)
		return 0, fmt.Errorf("failed to calculate total token amount: %w", err)
	}

	return totalTokenAmount, nil
}
