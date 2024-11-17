package token_transfer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
)

type tokenTransferUCase struct {
	tokenTransferRepository interfaces.TokenTransferRepository
	ethClient               *ethclient.Client
	config                  *conf.Configuration
}

func NewTokenTransferUCase(tokenTransferRepository interfaces.TokenTransferRepository, ethClient *ethclient.Client, config *conf.Configuration) interfaces.TokenTransferUCase {
	return &tokenTransferUCase{
		tokenTransferRepository: tokenTransferRepository,
		ethClient:               ethClient,
		config:                  config,
	}
}

// TransferTokens handles the entire process of token transfers and returns the status of each transaction
func (u *tokenTransferUCase) TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) ([]dto.TokenTransferResultDTOResponse, error) {
	// Initialize a slice to track the status of each transfer
	var transferResults []dto.TokenTransferResultDTOResponse

	// Group payloads by FromAddress and Symbol
	groupedPayloads := u.groupPayloadsByFromAddressAndSymbol(payloads)

	for key, grouped := range groupedPayloads {
		fromAddress, symbol := key.FromAddress, key.Symbol

		recipients, amounts, err := u.convertToRecipientsAndAmounts(grouped)
		if err != nil {
			// Log the error and mark each request in the group as failed
			for _, payload := range grouped {
				transferResults = append(transferResults, dto.TokenTransferResultDTOResponse{
					RequestID:    payload.RequestID,
					Status:       false,
					ErrorMessage: fmt.Sprintf("Failed to convert to recipients and amounts: %v", err),
				})
			}
			continue
		}

		tokenTransfers, err := u.prepareTokenTransferHistories(grouped)
		if err != nil {
			// Log the error and mark each request in the group as failed
			for _, payload := range grouped {
				transferResults = append(transferResults, dto.TokenTransferResultDTOResponse{
					RequestID:    payload.RequestID,
					Status:       false,
					ErrorMessage: fmt.Sprintf("Failed to prepare token transfer histories: %v", err),
				})
			}
			continue
		}

		tokenTransfers, err = u.bulkTransferAndSaveTokenTransferHistories(ctx, tokenTransfers, fromAddress, symbol, recipients, amounts)
		if err != nil {
			return nil, fmt.Errorf("failed to TransferTokens: %v", err)
		}
		for _, tokenTransfer := range tokenTransfers {
			transferResults = append(transferResults, dto.TokenTransferResultDTOResponse{
				RequestID:    tokenTransfer.RequestID,
				Status:       tokenTransfer.Status,
				ErrorMessage: tokenTransfer.ErrorMessage,
			})
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
		multiplier := new(big.Float).SetFloat64(float64(constants.TokenDecimalsMultiplier))

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
func (u *tokenTransferUCase) prepareTokenTransferHistories(payloads []dto.TokenTransferPayloadDTO) ([]domain.TokenTransferHistory, error) {
	var tokenTransfers []domain.TokenTransferHistory

	for _, payload := range payloads {
		// Prepare reward entry
		tokenTransfers = append(tokenTransfers, domain.TokenTransferHistory{
			RequestID:   payload.RequestID,
			Network:     payload.Network,
			FromAddress: payload.FromAddress,
			ToAddress:   payload.ToAddress,
			TokenAmount: payload.TokenAmount,
			Status:      false, // Default to failed status initially
		})
	}

	return tokenTransfers, nil
}

// bulkTransferAndSaveTokenTransferHistories performs bulk token transfer and updates token transfer history.
func (u *tokenTransferUCase) bulkTransferAndSaveTokenTransferHistories(
	ctx context.Context,
	tokenTransfers []domain.TokenTransferHistory,
	fromAddress, symbol string,
	recipients []string,
	amounts []*big.Int,
) ([]domain.TokenTransferHistory, error) {
	// TODO: currently this supports only AVAX C-Chain, need support other networks like BSC
	chainID := u.config.Blockchain.AvaxNetwork.AvaxChainID
	bulkSenderContractAddress := u.config.Blockchain.AvaxNetwork.AvaxBulkSenderContractAddress
	poolPrivateKey, err := u.config.GetPoolPrivateKey(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key for pool address: %s, %w", fromAddress, err)
	}
	var tokenContractAddress string
	if symbol == constants.USDT {
		tokenContractAddress = u.config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress
	} else {
		tokenContractAddress = u.config.Blockchain.AvaxNetwork.AvaxLifePointContractAddress
	}

	// Perform the bulk transfer.
	txHash, tokenSymbol, txFee, err := blockchain.BulkTransfer(
		ctx,
		u.ethClient,
		uint64(chainID),
		bulkSenderContractAddress,
		fromAddress,
		poolPrivateKey,
		tokenContractAddress,
		recipients,
		amounts,
	)

	// Process the transfer results.
	for index := range tokenTransfers {
		if err != nil {
			u.handleTransferError(&tokenTransfers[index], err, txHash, tokenSymbol, txFee)
		} else {
			// Handle success for the first token transfer, using actual txFee
			if index == 0 {
				u.handleTransferSuccess(&tokenTransfers[index], txHash, tokenSymbol, txFee)
			} else {
				zeroFee := big.NewFloat(0) // Set fee as zero for subsequent token transfers in transaction
				u.handleTransferSuccess(&tokenTransfers[index], txHash, tokenSymbol, zeroFee)
			}
		}
	}

	// Save the updated token transfer histories.
	if saveErr := u.tokenTransferRepository.CreateTokenTransferHistories(ctx, tokenTransfers); saveErr != nil {
		return nil, fmt.Errorf("failed to save token transfer histories: %v", saveErr)
	}

	return tokenTransfers, nil
}

// handleTransferSuccess updates token transfer history on success.
func (u *tokenTransferUCase) handleTransferSuccess(
	tokenTransfer *domain.TokenTransferHistory,
	txHash *string,
	tokenSymbol *string,
	txFee *big.Float,
) {
	tokenTransfer.TransactionHash = *txHash
	tokenTransfer.Status = true
	tokenTransfer.Symbol = *tokenSymbol
	tokenTransfer.Fee = txFee.String()
}

// handleTransferError updates token transfer history on error.
func (u *tokenTransferUCase) handleTransferError(
	tokenTransfer *domain.TokenTransferHistory,
	err error,
	txHash *string,
	tokenSymbol *string,
	txFee *big.Float,
) {
	tokenTransfer.Status = false
	tokenTransfer.ErrorMessage = err.Error()

	if tokenSymbol != nil {
		tokenTransfer.Symbol = *tokenSymbol
	}

	if txHash != nil {
		tokenTransfer.TransactionHash = *txHash
	}

	tokenTransfer.Fee = "0"
	if txFee != nil {
		tokenTransfer.Fee = txFee.String()
	}
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
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1
	offset := (page - 1) * size

	// Fetch the token transfer histories using request IDs and time range
	listTokenTransfers, err := s.tokenTransferRepository.GetTokenTransferHistories(ctx, limit, offset, requestIDs, startTime, endTime)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	var listTokenTransfersDTO []interface{}

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
