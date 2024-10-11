package token_transfer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	"github.com/genefriendway/onchain-handler/utils"
)

type tokenTransferUCase struct {
	TokenTransferRepository interfaces.TokenTransferRepository
	ETHClient               *ethclient.Client
	Config                  *conf.Configuration
}

func NewTokenTransferUCase(tokenTransferRepository interfaces.TokenTransferRepository, ethClient *ethclient.Client, config *conf.Configuration) interfaces.TokenTransferUCase {
	return &tokenTransferUCase{
		TokenTransferRepository: tokenTransferRepository,
		ETHClient:               ethClient,
		Config:                  config,
	}
}

// TransferTokens handles the entire process of tokens transfer
func (u *tokenTransferUCase) TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) error {
	// Convert the payload into two lists: one for recipients and one for amounts
	recipients, amounts, err := u.convertToRecipientsAndAmounts(payloads)
	if err != nil {
		return fmt.Errorf("failed to convert recipients: %v", err)
	}

	// Prepare token transfer history
	tokenTransfers, err := u.prepareTokenTransferHistories(payloads)
	if err != nil {
		return fmt.Errorf("failed to prepare token transfer histories: %v", err)
	}

	// Perform bulk token transfer
	err = u.bulkTransferAndSaveTokenTransferHistories(ctx, tokenTransfers, payloads[0].FromAddress, payloads[0].Symbol, recipients, amounts)
	if err != nil {
		return fmt.Errorf("failed to distribute tokens: %v", err)
	}

	return nil
}

// convertToRecipientsAndAmounts converts the payload into two lists: recipients and amounts (token amounts in smallest unit)
func (u *tokenTransferUCase) convertToRecipientsAndAmounts(req []dto.TokenTransferPayloadDTO) ([]string, []*big.Int, error) {
	var recipients []string
	var amounts []*big.Int

	for _, payload := range req {
		// Convert token amount to big.Float to handle fractional amounts
		tokenAmount := new(big.Float)
		_, ok := tokenAmount.SetString(payload.TokenAmount)
		if !ok {
			return nil, nil, fmt.Errorf("invalid token amount: %s", payload.TokenAmount)
		}

		// Create a multiplier for 10^18
		multiplier := new(big.Float).SetFloat64(float64(1e18)) // Adjust this based on your token's decimal places

		// Multiply the token amount by the multiplier
		tokenAmountInSmallestUnitFloat := new(big.Float).Mul(tokenAmount, multiplier)

		// Convert the big.Float amount to big.Int (smallest unit for blockchain transfer)
		tokenAmountInSmallestUnit := new(big.Int)
		tokenAmountInSmallestUnitFloat.Int(tokenAmountInSmallestUnit) // Truncate the fractional part

		// Append recipient address and token amount to their respective lists
		recipients = append(recipients, payload.ToAddress)
		amounts = append(amounts, tokenAmountInSmallestUnit)
	}

	return recipients, amounts, nil
}

// prepareTokenTransferHistories prepares token transfer history based on the payload
func (u *tokenTransferUCase) prepareTokenTransferHistories(req []dto.TokenTransferPayloadDTO) ([]model.TokenTransferHistory, error) {
	var tokenTransfers []model.TokenTransferHistory

	for _, payload := range req {
		// Prepare reward entry
		tokenTransfers = append(tokenTransfers, model.TokenTransferHistory{
			RequestID:   payload.RequestID,
			FromAddress: payload.FromAddress,
			ToAddress:   payload.ToAddress,
			TokenAmount: payload.TokenAmount,
			Status:      false, // Default to failed status initially
		})
	}

	return tokenTransfers, nil
}

// bulkTransferAndSaveTokenTransferHistories performs bulk token transfer and updates token transfer history
func (u *tokenTransferUCase) bulkTransferAndSaveTokenTransferHistories(
	ctx context.Context,
	tokenTransfers []model.TokenTransferHistory,
	fromAddress, symbol string,
	recipients []string,
	amounts []*big.Int,
) error {
	// Call the BulkTransfer utility to send tokens
	txHash, tokenSymbol, txFee, err := utils.BulkTransfer(u.ETHClient, u.Config, fromAddress, symbol, recipients, amounts)
	if err != nil {
		return fmt.Errorf("failed to bulk transfer token: %v", err)
	}
	for index := range tokenTransfers {
		// Update the token transfer history with transaction details
		tokenTransfers[index].TransactionHash = *txHash
		tokenTransfers[index].Status = true
		tokenTransfers[index].Symbol = *tokenSymbol
		tokenTransfers[index].Fee = txFee.String()
	}

	// Save the updated reward history
	err = u.TokenTransferRepository.CreateTokenTransferHistories(ctx, tokenTransfers)
	if err != nil {
		return fmt.Errorf("failed to save token transfer histories: %v", err)
	}

	return nil
}

// GetTokenTransferHistories fetches token transfer histories from the repository with optional filters
func (s *tokenTransferUCase) GetTokenTransferHistories(ctx context.Context, filters dto.TokenTransferFilterDTO, page, size int) (dto.TokenTransferHistoryDTOResponse, error) {
	// Convert filters DTO to a map
	filterMap := filters.ToMap()

	// Fetch the token transfer histories with filters
	limit := size + 1
	offset := (page - 1) * size
	listTokenTransfers, err := s.TokenTransferRepository.GetTokenTransferHistories(ctx, filterMap, limit, offset)
	if err != nil {
		return dto.TokenTransferHistoryDTOResponse{}, err
	}

	var listTokenTransfersDTO []dto.TokenTransferHistoryDTO

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
	return dto.TokenTransferHistoryDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     listTokenTransfersDTO,
	}, nil
}
