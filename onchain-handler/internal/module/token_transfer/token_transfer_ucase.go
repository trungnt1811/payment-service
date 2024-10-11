package token_transfer

import (
	"context"
	"fmt"
	"math/big"
	"time"

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

// TransferTokens handles the entire process of token transfers and returns the status of each transaction
func (u *tokenTransferUCase) TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) (map[string]string, error) {
	// Initialize a map to track the status of each transfer, with error details
	transferResults := make(map[string]string)

	// Group payloads by FromAddress and Symbol
	groupedPayloads := u.groupPayloadsByFromAddressAndSymbol(payloads)

	for key, grouped := range groupedPayloads {
		fromAddress, symbol := key.FromAddress, key.Symbol

		recipients, amounts, err := u.convertToRecipientsAndAmounts(grouped)
		if err != nil {
			// Log the error and mark the batch as failed, including the error message
			for _, payload := range grouped {
				transferResults[payload.RequestID] = fmt.Sprintf("Failed: %v", err)
			}
			continue
		}

		tokenTransfers, err := u.prepareTokenTransferHistories(grouped)
		if err != nil {
			// Log the error and mark the batch as failed, including the error message
			for _, payload := range grouped {
				transferResults[payload.RequestID] = fmt.Sprintf("Failed: %v", err)
			}
			continue
		}

		err = u.bulkTransferAndSaveTokenTransferHistories(ctx, tokenTransfers, fromAddress, symbol, recipients, amounts)
		if err != nil {
			// Log the error and mark the batch as failed, including the error message
			for _, payload := range grouped {
				transferResults[payload.RequestID] = fmt.Sprintf("Failed: %v", err)
			}
		} else {
			// Mark all transactions in the batch as successful
			for _, payload := range grouped {
				transferResults[payload.RequestID] = "Success"
			}
		}
	}

	return transferResults, nil
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

// groupKey is used to create a unique key for grouping by FromAddress and Symbol
type groupKey struct {
	FromAddress string
	Symbol      string
}

// groupPayloadsByFromAddressAndSymbol groups payloads by FromAddress and Symbol
func (u *tokenTransferUCase) groupPayloadsByFromAddressAndSymbol(payloads []dto.TokenTransferPayloadDTO) map[groupKey][]dto.TokenTransferPayloadDTO {
	groupedPayloads := make(map[groupKey][]dto.TokenTransferPayloadDTO)

	for _, payload := range payloads {
		key := groupKey{
			FromAddress: payload.FromAddress,
			Symbol:      payload.Symbol,
		}

		groupedPayloads[key] = append(groupedPayloads[key], payload)
	}

	return groupedPayloads
}

// GetTokenTransferHistories fetches token transfer histories from the repository with optional filters
func (s *tokenTransferUCase) GetTokenTransferHistories(
	ctx context.Context,
	requestIDs []string, // List of request IDs to filter
	startTime, endTime time.Time, // Time range to filter by
	page, size int,
) (dto.TokenTransferHistoryDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1
	offset := (page - 1) * size

	// Fetch the token transfer histories using request IDs and time range
	listTokenTransfers, err := s.TokenTransferRepository.GetTokenTransferHistories(ctx, limit, offset, requestIDs, startTime, endTime)
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
