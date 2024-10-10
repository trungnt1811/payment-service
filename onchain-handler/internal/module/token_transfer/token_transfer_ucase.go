package transfer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
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

	// Prepare reward history
	rewardModels, err := u.prepareTokenTransferHistories(payloads)
	if err != nil {
		return fmt.Errorf("failed to prepare token transfer histories: %v", err)
	}

	// Perform bulk token transfer
	err = u.bulkTransferAndSaveTokenTransferHistories(ctx, rewardModels, recipients, amounts)
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
		// Convert token amount to big.Int
		tokenAmount := new(big.Int)
		tokenAmount.SetUint64(payload.TokenAmount)

		// Multiply by 10^18 to convert to the smallest unit of the token (like wei for ETH)
		tokenAmountInSmallestUnit := new(big.Int).Mul(tokenAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(constants.LifePointDecimals), nil))

		// Append recipient address and token amount to their respective lists
		recipients = append(recipients, payload.ToAddress)
		amounts = append(amounts, tokenAmountInSmallestUnit)
	}

	return recipients, amounts, nil
}

// prepareTokenTransferHistories prepares token transfer history based on the payload
func (u *tokenTransferUCase) prepareTokenTransferHistories(req []dto.TokenTransferPayloadDTO) ([]model.TokenTransferHistory, error) {
	var rewards []model.TokenTransferHistory

	for _, payload := range req {
		// Prepare reward entry
		rewards = append(rewards, model.TokenTransferHistory{
			RequestID:   payload.RequestID,
			FromAddress: u.Config.Blockchain.LPTreasuryPool.LPTreasuryAddress,
			ToAddress:   payload.ToAddress,
			TokenAmount: payload.TokenAmount,
			Status:      false, // Default to failed status initially
		})
	}

	return rewards, nil
}

// bulkTransferAndSaveTokenTransferHistories performs bulk token transfer and updates token transfer history
func (u *tokenTransferUCase) bulkTransferAndSaveTokenTransferHistories(
	ctx context.Context,
	tokenTransfers []model.TokenTransferHistory,
	recipients []string,
	amounts []*big.Int,
) error {
	// Call the BulkTransfer utility to send tokens
	txHash, tokenSymbol, txFee, err := utils.BulkTransfer(u.ETHClient, u.Config, u.Config.Blockchain.LPTreasuryPool.LPTreasuryAddress, recipients, amounts)
	for index := range tokenTransfers {
		if err != nil {
			return fmt.Errorf("failed to bulk transfer token: %v", err)
		} else {
			// Update the token transfer history with transaction details
			tokenTransfers[index].TransactionHash = *txHash
			tokenTransfers[index].Status = true
			tokenTransfers[index].Symbol = *tokenSymbol
			fee, _ := txFee.Float64()
			tokenTransfers[index].Fee = fee
		}
	}

	// Save the updated reward history
	err = u.TokenTransferRepository.CreateTokenTransferHistories(ctx, tokenTransfers)
	if err != nil {
		return fmt.Errorf("failed to save token transfer histories: %v", err)
	}

	return nil
}
