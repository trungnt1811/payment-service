package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/contracts/abigen/lifepointtoken"
	util "github.com/genefriendway/onchain-handler/internal/utils/ethereum"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

// DistributeTokens distributes tokens from the token distribution address to user wallets using bulk transfer
func DistributeTokens(client *ethclient.Client, config *conf.Configuration, recipients map[string]*big.Int) (*string, error) {
	// Load Blockchain configuration
	chainID := config.Blockchain.ChainID
	privateKey := config.Blockchain.PrivateKeyReward
	tokenAddress := config.Blockchain.LifePointAddress

	// Get authentication for signing transactions
	privateKeyECDSA, err := util.PrivateKeyFromHex(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	auth, err := util.GetAuth(client, privateKeyECDSA, new(big.Int).SetUint64(uint64(chainID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get auth: %w", err)
	}

	// Set up the reward token contract instance
	LPToken, err := lifepointtoken.NewLifepointtoken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ERC20 contract: %w", err)
	}

	// Prepare recipient addresses and values for bulk transfer
	var recipientAddresses []common.Address
	var tokenAmounts []*big.Int
	for recipientAddress, amount := range recipients {
		recipientAddresses = append(recipientAddresses, common.HexToAddress(recipientAddress))
		tokenAmounts = append(tokenAmounts, amount)
	}

	// Call the bulkTransfer function in the Solidity contract
	tx, err := LPToken.BulkTransfer(auth, recipientAddresses, tokenAmounts)
	if err != nil {
		log.LG.Errorf("Failed to execute bulk transfer: %v", err)
		return nil, err
	}

	// Get the transaction hash after a successful transfer
	txHash := tx.Hash().Hex()

	// Log the transaction hash for tracking
	log.LG.Infof("Bulk transfer executed. Tx hash: %s\n", txHash)

	// Return success with the transaction hash
	return &txHash, nil
}
