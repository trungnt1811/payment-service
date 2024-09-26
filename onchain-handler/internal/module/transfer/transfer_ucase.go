package transfer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/blockchain"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type transferUCase struct {
	TrasferRepository interfaces.TransferRepository
	ETHClient         *ethclient.Client
	Config            *conf.Configuration
}

func NewtTransferUCase(transferRepository interfaces.TransferRepository, ethClient *ethclient.Client, config *conf.Configuration) interfaces.TransferUCase {
	return &transferUCase{
		TrasferRepository: transferRepository,
		ETHClient:         ethClient,
		Config:            config,
	}
}

// DistributeTokens handles the entire process of tokens distribution
func (u *transferUCase) DistributeTokens(ctx context.Context, payloads []dto.TransferTokenPayloadDTO) error {
	// Convert the payload into recipients
	recipients, err := u.convertToRecipients(payloads)
	if err != nil {
		return fmt.Errorf("failed to convert recipients: %v", err)
	}

	// Prepare reward history
	rewardModels, err := u.prepareRewardHistory(payloads)
	if err != nil {
		return fmt.Errorf("failed to prepare reward history: %v", err)
	}

	// Perform concurrent reward distribution
	err = u.distributeAndSaveRewards(ctx, rewardModels, recipients)
	if err != nil {
		return fmt.Errorf("failed to distribute rewards: %v", err)
	}

	return nil
}

// convertToRecipients converts the payload into recipients (address -> token amount in smallest unit)
func (u *transferUCase) convertToRecipients(req []dto.TransferTokenPayloadDTO) (map[string]*big.Int, error) {
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

// prepareRewardHistory prepares reward history based on the payload
func (u *transferUCase) prepareRewardHistory(req []dto.TransferTokenPayloadDTO) ([]model.TransferHistory, error) {
	var rewards []model.TransferHistory

	for _, payload := range req {
		// Validate token amount
		tokenAmount := new(big.Int)
		if _, success := tokenAmount.SetString(payload.TokenAmount, 10); !success {
			return nil, fmt.Errorf("invalid token amount: %s", payload.TokenAmount)
		}

		// Prepare reward entry
		rewards = append(rewards, model.TransferHistory{
			RewardAddress:    u.Config.Blockchain.RewardAddress,
			RecipientAddress: payload.RecipientAddress,
			TokenAmount:      payload.TokenAmount,
			Status:           -1, // Default to failed status initially
			TxType:           payload.TxType,
		})
	}

	return rewards, nil
}

// distributeAndSaveRewards distributes rewards and updates reward history
func (u *transferUCase) distributeAndSaveRewards(ctx context.Context, rewards []model.TransferHistory, recipients map[string]*big.Int) error {
	txHash, err := blockchain.DistributeReward(u.ETHClient, u.Config, recipients)
	for index := range rewards {
		if err != nil {
			rewards[index].ErrorMessage = fmt.Sprintf("Failed to distribute: %v", err)
			rewards[index].Status = -1
		} else {
			rewards[index].TransactionHash = *txHash
			rewards[index].Status = 1
		}
	}

	// Save reward history
	err = u.TrasferRepository.CreateTransferHistories(ctx, rewards)
	if err != nil {
		return fmt.Errorf("failed to save rewards history: %v", err)
	}

	return nil
}
