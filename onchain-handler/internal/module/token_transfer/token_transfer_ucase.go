package transfer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	"github.com/genefriendway/onchain-handler/utils"
)

type tokenTransferUCase struct {
	TokenTrasferRepository interfaces.TokenTransferRepository
	ETHClient              *ethclient.Client
	Config                 *conf.Configuration
}

func NewTokenTransferUCase(tokenTransferRepository interfaces.TokenTransferRepository, ethClient *ethclient.Client, config *conf.Configuration) interfaces.TokenTransferUCase {
	return &tokenTransferUCase{
		TokenTrasferRepository: tokenTransferRepository,
		ETHClient:              ethClient,
		Config:                 config,
	}
}

// TransferTokens handles the entire process of tokens transfer
func (u *tokenTransferUCase) TransferTokens(ctx context.Context, payloads []dto.TokenTransferPayloadDTO) error {
	// Convert the payload into recipients
	recipients, err := u.convertToRecipients(payloads)
	if err != nil {
		return fmt.Errorf("failed to convert recipients: %v", err)
	}

	// Prepare reward history
	rewardModels, err := u.prepareTransferHistories(payloads)
	if err != nil {
		return fmt.Errorf("failed to prepare token transfer histories: %v", err)
	}

	// Perform concurrent tokens distribution
	err = u.distributeAndSaveTransferHistories(ctx, rewardModels, recipients)
	if err != nil {
		return fmt.Errorf("failed to distribute tokens: %v", err)
	}

	return nil
}

// convertToRecipients converts the payload into recipients (address -> token amount in smallest unit)
func (u *tokenTransferUCase) convertToRecipients(req []dto.TokenTransferPayloadDTO) (map[string]*big.Int, error) {
	recipients := make(map[string]*big.Int)

	for _, payload := range req {
		// Check for duplicate recipient addresses
		if _, exists := recipients[payload.RecipientAddress]; exists {
			return nil, fmt.Errorf("duplicate recipient address: %s", payload.RecipientAddress)
		}

		// Convert token amount to big.Int
		tokenAmount := new(big.Int)
		if _, success := tokenAmount.SetString(payload.TokenAmount, 10); !success {
			return nil, fmt.Errorf("invalid token amount: %s", payload.TokenAmount)
		}

		// Multiply by 10^18 to convert to the smallest unit of the token (like wei for ETH)
		tokenAmountInSmallestUnit := new(big.Int).Mul(tokenAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(constants.LifePointDecimals), nil))
		recipients[payload.RecipientAddress] = tokenAmountInSmallestUnit
	}

	return recipients, nil
}

// prepareTransferHistories prepares token transfer history based on the payload
func (u *tokenTransferUCase) prepareTransferHistories(req []dto.TokenTransferPayloadDTO) ([]model.TokenTransferHistory, error) {
	var rewards []model.TokenTransferHistory

	for _, payload := range req {
		// Validate token amount
		tokenAmount := new(big.Int)
		if _, success := tokenAmount.SetString(payload.TokenAmount, 10); !success {
			return nil, fmt.Errorf("invalid token amount: %s", payload.TokenAmount)
		}

		// Prepare reward entry
		rewards = append(rewards, model.TokenTransferHistory{
			FromAddress: u.Config.Blockchain.LPTreasuryPool.LPTreasuryAddress,
			ToAddress:   payload.RecipientAddress,
			TokenAmount: payload.TokenAmount,
			Status:      false, // Default to failed status initially
		})
	}

	return rewards, nil
}

// distributeAndSaveRewards distributes rewards and updates reward history
func (u *tokenTransferUCase) distributeAndSaveTransferHistories(ctx context.Context, rewards []model.TokenTransferHistory, recipients map[string]*big.Int) error {
	txHash, _, err := utils.BulkTransfer(u.ETHClient, u.Config, u.Config.Blockchain.LPTreasuryPool.LPTreasuryAddress, recipients)
	for index := range rewards {
		if err != nil {
			rewards[index].ErrorMessage = fmt.Sprintf("Failed to distribute: %v", err)
			rewards[index].Status = false
		} else {
			rewards[index].TransactionHash = *txHash
			rewards[index].Status = false
		}
	}

	// Save reward history
	err = u.TokenTrasferRepository.CreateTokenTransferHistories(ctx, rewards)
	if err != nil {
		return fmt.Errorf("failed to save token transfer histories: %v", err)
	}

	return nil
}
