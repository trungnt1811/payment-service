package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/blockchain/interfaces"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/contracts/abigen/bulksender"
	"github.com/genefriendway/onchain-handler/contracts/abigen/erc20token"
)

// BulkTransfer transfers tokens from the pool address to recipients using bulk transfer
func BulkTransfer(
	ctx context.Context,
	client *ethclient.Client,
	config *conf.Configuration,
	poolAddress, symbol string,
	recipients []string,
	amounts []*big.Int,
) (*string, *string, *big.Float, error) {
	chainID := config.Blockchain.ChainID
	bulkSenderContractAddress := config.Blockchain.SmartContract.BulkSenderContractAddress

	var txHash, tokenSymbol string
	var txFeeInAVAX *big.Float

	// Get token address, pool private key, and symbol based on the symbol
	var erc20Token interface{}
	var err error
	var tokenAddress, poolPrivateKey string
	// Get pool private key
	poolPrivateKey, err = config.GetPoolPrivateKey(poolAddress)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to get private key for pool address: %s", poolAddress)
	}
	// Get erc20 token
	if symbol == constants.USDT {
		tokenAddress = config.Blockchain.SmartContract.USDTContractAddress
	} else {
		tokenAddress = config.Blockchain.SmartContract.LifePointContractAddress
	}
	erc20Token, err = erc20token.NewErc20token(common.HexToAddress(tokenAddress), client)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to instantiate ERC20 contract for %s: %w", symbol, err)
	}

	privateKeyECDSA, err := PrivateKeyFromHex(poolPrivateKey)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve pool private key: %w", err)
	}

	// Function to handle nonce retrieval and retry logic
	var nonce uint64
	nonce, err = getNonceWithRetry(ctx, client, poolAddress)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve nonce after retry: %w", err)
	}

	auth, err := getAuth(ctx, client, privateKeyECDSA, new(big.Int).SetUint64(uint64(chainID)))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to create auth object for pool %s: %w", poolAddress, err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce) // Set the correct nonce

	// Set up the bulk transfer contract instance
	bulkSender, err := bulksender.NewBulksender(common.HexToAddress(bulkSenderContractAddress), client)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to instantiate bulk sender contract: %w", err)
	}

	// Type assertion to ERC20Token interface
	token, ok := erc20Token.(interfaces.ERC20Token)
	if !ok {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("erc20Token does not implement ERC20Token interface for %s", symbol)
	}

	// Get the token symbol from the contract
	tokenSymbol, err = token.Symbol(nil)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve token symbol for %s: %w", symbol, err)
	}

	// Calculate total amount to transfer for approval
	totalAmount := big.NewInt(0)
	for _, amount := range amounts {
		totalAmount.Add(totalAmount, amount)
	}

	// Check pool address balance
	poolBalance, err := token.BalanceOf(nil, common.HexToAddress(poolAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to get pool balance: %w", err)
	}

	// Ensure the pool has enough balance
	if poolBalance.Cmp(totalAmount) < 0 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("insufficient pool balance: required %s %s, available %s %s", totalAmount.String(), tokenSymbol, poolBalance.String(), tokenSymbol)
	}

	// Approve the bulk transfer contract to spend tokens on behalf of the pool wallet
	tx, err := token.Approve(auth, common.HexToAddress(bulkSenderContractAddress), totalAmount)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to approve bulk sender contract: %w", err)
	}
	txHash = tx.Hash().Hex() // Get the transaction hash

	// Wait for approval to be mined
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to wait for approval transaction to be mined: %w", err)
	}
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("approval transaction failed for %s", txHash)
	}

	// Increment nonce for the next transaction
	nonce++
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Call the bulk transfer function on the bulk sender contract
	tx, err = bulkSender.BulkTransfer(auth, convertToCommonAddresses(recipients), amounts, common.HexToAddress(tokenAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to execute bulk transfer: %w", err)
	}
	txHash = tx.Hash().Hex() // Update transaction hash for bulk transfer

	// Wait for the bulk transfer transaction to be mined
	receipt, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to wait for bulk transfer transaction to be mined: %w", err)
	}

	// Calculate transaction fee (gasUsed * gasPrice)
	gasUsed := receipt.GasUsed
	gasPrice := auth.GasPrice
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)
	// Convert txFee from wei to AVAX (1 AVAX = 10^18 wei)
	weiInAVAX := big.NewFloat(constants.TokenDecimalsMultiplier)
	txFeeInAVAX = new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInAVAX)

	// Check the transaction status
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
	}

	return &txHash, &tokenSymbol, txFeeInAVAX, nil
}

// getAuth creates a new keyed transactor for signing transactions with the given private key and network chain ID
func getAuth(
	ctx context.Context,
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
) (*bind.TransactOpts, error) {
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // 0 wei, since we're not sending Ether
	// auth.GasLimit = uint64(300000) // Set the gas limit (adjust as needed)
	auth.GasPrice = gasPrice // Set the gas price

	return auth, nil
}

// Helper function to convert string addresses to common.Address type
func convertToCommonAddresses(recipients []string) []common.Address {
	var addresses []common.Address
	for _, recipient := range recipients {
		addresses = append(addresses, common.HexToAddress(recipient))
	}
	return addresses
}

// Helper function to retry nonce retrieval
func getNonceWithRetry(ctx context.Context, client *ethclient.Client, poolAddress string) (uint64, error) {
	var nonce uint64
	var err error
	for retryCount := 0; retryCount < constants.MaxRetries; retryCount++ {
		nonce, err = client.PendingNonceAt(ctx, common.HexToAddress(poolAddress))
		if err == nil {
			return nonce, nil
		}
		time.Sleep(constants.RetryDelay) // Backoff before retrying
	}
	return 0, fmt.Errorf("failed to retrieve nonce after %d retries: %w", constants.MaxRetries, err)
}

// PrivateKeyFromHex converts a private key string in hex format to an ECDSA private key
func PrivateKeyFromHex(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key from hex: %w", err)
	}
	return privateKey, nil
}
